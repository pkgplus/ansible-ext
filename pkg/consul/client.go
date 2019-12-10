package consul

import (
	consulapi "github.com/hashicorp/consul/api"
)

func getClient(consulAddress string, dataCenter string) (*consulapi.Client, error) {
	config := consulapi.DefaultConfig()
	config.Address = consulAddress
	if dataCenter != "" {
		config.Datacenter = dataCenter
	}
	client, err := consulapi.NewClient(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
