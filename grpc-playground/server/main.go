package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	userpb "github.com/cristianmanoliu/learning-golang/grpc-playground/proto_gen/proto"
)

type userServer struct {
	userpb.UnimplementedUserServiceServer

	mu    sync.Mutex
	users []*userpb.User
}

func (s *userServer) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
	if req.GetName() == "" {
		return nil, fmt.Errorf("name is required")
	}

	u := &userpb.User{
		Id:   uuid.NewString(),
		Name: req.GetName(),
	}

	s.mu.Lock()
	s.users = append(s.users, u)
	s.mu.Unlock()

	log.Printf("CreateUser: id=%s name=%s\n", u.Id, u.Name)

	return &userpb.CreateUserResponse{User: u}, nil
}

func (s *userServer) ListUsers(ctx context.Context, _ *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	usersCopy := make([]*userpb.User, len(s.users))
	copy(usersCopy, s.users)

	log.Printf("ListUsers: returning %d users\n", len(usersCopy))

	return &userpb.ListUsersResponse{Users: usersCopy}, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	userpb.RegisterUserServiceServer(grpcServer, &userServer{})

	log.Println("gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Serve: %v", err)
	}
}
