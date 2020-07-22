package main

import (
	"net"
	"todo-app/todo"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
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

func main() {
	source := "root:1@tcp(127.0.0.1:3306)/todoapp"
	db := sqlx.MustConnect("mysql", source)

	service := todo.NewService(db)
	runService(service)
}
