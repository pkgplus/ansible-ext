package main

import (
	"io"
	"log"
	"time"

	pb "github.com/xuebing1110/ansible-ext/proto/ansible"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address = "10.138.16.192:50051"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewAnsibleClient(conn)

	// // precheck
	// r, err := c.Precheck(context.Background(),
	// 	&pb.PrecheckRequest{
	// 		LoginInfos: []*pb.LoginInfo{
	// 			&pb.LoginInfo{
	// 				Host:     "10.138.16.188",
	// 				Port:     22,
	// 				UserName: "haieradmin",
	// 				Passwd:   "123,Haier",
	// 			},
	// 		},
	// 	},
	// )
	// if err != nil {
	// 	log.Fatalf("could not call precheck: %v", err)
	// }
	// log.Printf("Got precheck message %s[%s]", r.Results[0].Status, r.Results[0].Message)

	// // init
	// r2, err := c.InitHosts(context.Background(),
	// 	&pb.InitRequest{
	// 		LoginInfos: []*pb.LoginInfo{
	// 			&pb.LoginInfo{
	// 				Host:     "10.138.16.188",
	// 				Port:     22,
	// 				UserName: "haieradmin",
	// 				Passwd:   "123,Haier",
	// 			},
	// 		},
	// 	},
	// )
	// if err != nil {
	// 	log.Fatalf("could not call init: %v", err)
	// }
	// log.Printf("Got init message %s[%s]", r2.Results[0].Status, r2.Results[0].Message)

	// install
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := c.Play(ctx,
		&pb.PlayBook{
			Labels: map[string]string{},
			Items: []*pb.PlayBookItems{
				{
					Host:  "10.138.16.192",
					Names: []string{"node_exporter"},
				},
			},
		},
	)
	if err != nil {
		log.Fatalf("install failed: %v", err)
	}

	// READ_LOOP:
	for {
		msg, err := stream.Recv()
		if err != nil {
			if err != io.EOF {
				log.Printf("Got install error:%v", err)
			}
			break
		}

		if msg != nil {
			log.Printf("Got install message: %+v", msg)
		}
	}
}
