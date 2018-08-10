package consul

import (
	"errors"
	"fmt"

	"github.com/hashicorp/consul/api"
)

var defaultListenPort = map[string]int32{
	"node_exporter":  9100,
	"redis_exporter": 9121,
	"mysql_exporter": 9104,
}

//func getDefaultListenPort(path string) int32 {
//	paths := strings.Split(path, "/")
//	filename := strings.TrimSuffix(paths[len(paths)-1], ".yml")
//	return defaultListenPort[filename]
//}

func RegisterService(srvname string, host string, port int32, labelPairs map[string]string) error {
	if consulClient == nil {
		return errors.New("init consul client failed")
	}

	if port <= 0 {
		var found bool
		port, found = defaultListenPort[srvname]
		if !found {
			return errors.New("can't get listen port for service " + srvname)
		}
	}

	// service tags : ["labelname1=labelvalue1,labelname2=labelvalue2"]
	var tags = make([]string, 0, len(labelPairs))
	for name, value := range labelPairs {
		tags = append(tags, name+"="+value)
	}

	service := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%s-%d", srvname, host, port),
		Name:    srvname,
		Tags:    tags,
		Port:    int(port),
		Address: host,
		//Check: &api.AgentServiceCheck{
		//	HTTP:     fmt.Sprintf("http://%s:%d%s", host, port, pbi.CheckPath),
		//	Interval: "300s",
		//},
	}
	return consulClient.Agent().ServiceRegister(service)
}
