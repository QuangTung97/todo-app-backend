package main

import (
	"context"
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

func runGateway(gateway *todo.Gateway) {
	ctx := context.Background()
	r := mux.NewRouter()
	grpcRouter := runtime.NewServeMux()

	configGateway(ctx, grpcRouter)

	r.Handle("/accounts",
		gateway.Authenticated(grpcRouter)).Methods(http.MethodPost)

	srv := &http.Server{
		Handler:      r,
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

	source := "root:1@tcp(127.0.0.1:3306)/todoapp"
	db := sqlx.MustConnect("mysql", source)
	redisClient := connectToRedis()

	service := todo.NewService(db)
	go runService(service)

	gateway := todo.NewGateway(db, redisClient)
	runGateway(gateway)
}
