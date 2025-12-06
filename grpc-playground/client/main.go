package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userpb "github.com/cristianmanoliu/learning-golang/grpc-playground/proto_gen/proto"
)

func main() {
	// Add timestamps + file:line to logs
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Dial the gRPC server running on localhost:50051
	conn, err := grpc.Dial(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Create a typed client for UserService
	client := userpb.NewUserServiceClient(conn)

	// Use a context with timeout for both calls
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1) Call CreateUser
	createResp, err := client.CreateUser(ctx, &userpb.CreateUserRequest{
		Name: "Cristi",
	})
	if err != nil {
		log.Fatalf("CreateUser: %v", err)
	}

	fmt.Printf("Created user: id=%s name=%s\n",
		createResp.GetUser().GetId(),
		createResp.GetUser().GetName(),
	)

	// 2) Call ListUsers
	listResp, err := client.ListUsers(ctx, &userpb.ListUsersRequest{})
	if err != nil {
		log.Fatalf("ListUsers: %v", err)
	}

	fmt.Println("Users from server:")
	for _, u := range listResp.GetUsers() {
		fmt.Printf("  id=%s name=%s\n", u.GetId(), u.GetName())
	}
}
