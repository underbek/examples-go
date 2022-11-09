package server

import (
	"context"

	"github.com/docker/distribution/uuid"
	pb "github.com/underbek/examples-go/grpc/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userService struct {
	// Need for implementation interface pb.UserServiceServer
	pb.UnimplementedUserServiceServer
	users map[string]*pb.User
}

func New() *userService {
	return &userService{
		users: make(map[string]*pb.User),
	}
}

func (s *userService) CreateUser(ctx context.Context, request *pb.CreateUserRequest) (*pb.User, error) {

	user := &pb.User{
		Id:   uuid.Generate().String(),
		Name: request.Name,
	}

	if request.Email != nil {
		user.Email = *request.Email
	}

	s.users[user.Id] = user

	return user, nil
}

func (s *userService) GetUser(ctx context.Context, request *pb.GetUserRequest) (*pb.User, error) {
	user, ok := s.users[request.Id]
	if !ok {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"User with id \"%s\" not found",
			request.Id,
		)
	}

	return user, nil
}
