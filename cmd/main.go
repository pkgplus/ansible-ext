package main

import (
	"context"
	"log"
	"net"
	"os"

	"ansible-ext/gateway"
	"ansible-ext/proto/ansible"
	"ansible-ext/proto/hostmanager"
	"ansible-ext/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var GRPC_PORT string
var RESTFUL_FLAG string

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	GRPC_PORT = os.Getenv("GRPC_PORT")
	if GRPC_PORT == "" {
		GRPC_PORT = "50051"
	}

	RESTFUL_FLAG = os.Getenv("RESTFUL_FLAG")
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var exit = make(chan bool, 1)

	// grpc server
	go func() {
		defer func() {
			exit <- true
		}()

		log.Print("start grpc server...")
		err := grpcServer()
		if err != nil {
			log.Fatalf("failed to start grpc server: %v", err)
		}
	}()

	// rest server
	if RESTFUL_FLAG == "1" || RESTFUL_FLAG == "true" {
		go func() {
			defer func() {
				exit <- true
			}()

			log.Print("start rest server...")
			err := restfulServer(ctx)
			if err != nil {
				log.Fatalf("failed to start rest server: %v", err)
			}
		}()
	}

	// wait
	<-exit
}

func grpcServer() error {
	lis, err := net.Listen("tcp", ":"+GRPC_PORT)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	Ansible.RegisterAnsibleServer(s, new(server.AnsibleServer))
	HostManager.RegisterHostManagerServer(s, new(server.HostManagerServer))

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		return err
	}

	return nil
}

func restfulServer(ctx context.Context) error {
	grpc_server_addr := "127.0.0.1:" + GRPC_PORT
	listen_addr := ":8080"
	return gateway.RestfulServer(ctx, grpc_server_addr, listen_addr)
}
