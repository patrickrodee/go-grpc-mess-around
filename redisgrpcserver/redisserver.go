package main

import (
	pb "../protos/messages"
	"context"
	"flag"
	"fmt"
	"github.com/go-redis/redis"
	"google.golang.org/grpc"
	"net"
)

var (
	port = flag.Int("port", 10000, "The server port")
)

type messageServer struct {
	Client *redis.Client
}

func newMessageServer(redisOptions *redis.Options) *messageServer {
	redisClient := redis.NewClient(redisOptions)
	return &messageServer{
		Client: redisClient,
	}
}

func (m *messageServer) GetMessage(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
	resp := pb.MessageResponse{}
	val, err := m.Client.Get(req.GetKey()).Result()
	if err == redis.Nil {
		resp.Value = ""
	} else if err != nil {
		return nil, err
	}

	resp.Value = val
	return &resp, nil
}

func (m *messageServer) SaveMessage(ctx context.Context, req *pb.SaveMessageRequest) (*pb.SaveMessageResponse, error) {
	err := m.Client.Set(req.GetKey(), req.GetValue(), 0).Err()
	if err != nil {
		return nil, err
	}

	resp := &pb.SaveMessageResponse{
		Ok: true,
	}
	return resp, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Redis server listening on port %d\n", *port)

	grpcServer := grpc.NewServer()
	pb.RegisterMessageServiceServer(grpcServer, newMessageServer(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}))
	grpcServer.Serve(lis)
}
