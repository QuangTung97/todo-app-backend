package main

import (
	"context"
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"time"
	"todo-app/todo"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/jmoiron/sqlx"
	"github.com/rs/cors"
	"google.golang.org/grpc"
)

func runService(service *todo.Service) {
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		glog.Fatal(err)
	}

	opts := []grpc.ServerOption{}

	server := grpc.NewServer(opts...)
	todo.RegisterTodoAppServer(server, service)

	err = server.Serve(listener)
	if err != nil {
		glog.Fatal(err)
	}
	glog.Flush()
}

func configGateway(ctx context.Context, mux *runtime.ServeMux) {
	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := todo.RegisterTodoAppHandlerFromEndpoint(
		ctx, mux, "localhost:9000", opts)
	if err != nil {
		glog.Fatal(err)
	}
}

func setupCORSConfig() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5000"},
		AllowedHeaders: []string{"Authorization", "Content-Type", "X-Auth-Token"},
		ExposedHeaders: []string{"X-Auth-Token"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost,
			http.MethodOptions, http.MethodPut, http.MethodDelete},
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Status string `json:"status"`
	}
	res := Response{Status: "ok"}
	err := json.NewEncoder(w).Encode(&res)
	if err != nil {
		glog.Fatal(err)
	}
}

func runGateway(gateway *todo.Gateway) {
	ctx := context.Background()

	r := mux.NewRouter()
	grpcRouter := runtime.NewServeMux()
	c := setupCORSConfig()

	configGateway(ctx, grpcRouter)

	// Setup routes
	r.Handle("/accounts", grpcRouter).
		Methods(http.MethodPost)

	r.Handle("/todos",
		gateway.Authenticated(grpcRouter)).
		Methods(http.MethodPost, http.MethodPut, http.MethodGet)

	r.Handle("/todos/{id}",
		gateway.Authenticated(grpcRouter)).
		Methods(http.MethodDelete)

	r.Handle("/todo-items",
		gateway.Authenticated(grpcRouter)).
		Methods(http.MethodPost)

	r.Handle("/todo-items/{todo_list_id}",
		gateway.Authenticated(grpcRouter)).
		Methods(http.MethodGet)

	r.Handle("/todo-items/completed",
		gateway.Authenticated(grpcRouter)).
		Methods(http.MethodPut)

	r.Handle("/todo-items/completed/{todo_list_id}",
		gateway.Authenticated(grpcRouter)).
		Methods(http.MethodDelete)

	r.Handle("/login",
		gateway.Authenticated(http.HandlerFunc(loginHandler))).
		Methods(http.MethodPost)

	srv := &http.Server{
		Handler:      c.Handler(r),
		Addr:         ":8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	glog.Fatal(srv.ListenAndServe())
	glog.Flush()
}

func connectToRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()
	err := client.Ping(ctx).Err()
	if err != nil {
		glog.Fatal(err)
	}

	return client
}

func main() {
	flag.Parse()

	source := "root:1@tcp(127.0.0.1:3306)/todoapp?parseTime=true"
	db := sqlx.MustConnect("mysql", source)
	redisClient := connectToRedis()

	service := todo.NewService(db)
	go runService(service)

	gateway := todo.NewGateway(db, redisClient)
	runGateway(gateway)
}
