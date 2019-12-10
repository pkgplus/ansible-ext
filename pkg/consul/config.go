package consul

import (
	"log"
	"os"

	"github.com/hashicorp/consul/api"
)

var (
	dataCenter   string
	consulClient *api.Client
)

func init() {
	config := api.DefaultConfig()
	addr := os.Getenv("CONSUL_ADDR")
	if addr != "" {
		config.Address = addr
	}

	var err error
	consulClient, err = api.NewClient(config)
	if err != nil {
		log.Printf("create consul client with %s error: %v\n", addr, err)
	}

	dataCenter = os.Getenv("CONSUL_DATACENTER")
	if dataCenter == "" {
		dataCenter = "dc1"
	}
}
