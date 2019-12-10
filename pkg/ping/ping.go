package ping

import (
	"github.com/tatsushid/go-fastping"
	"net"
	"sync"
	"time"
)

func GetStatus(hosts []string) ([]int, error) {
	p := fastping.NewPinger()
	ret := make([]int, len(hosts))
	retMap := make(map[string]time.Duration)

	wg := sync.WaitGroup{}
	wg.Add(1)
	for _, host := range hosts {
		ra, err := net.ResolveIPAddr("ip4:icmp", host)
		if err != nil {
			return nil, err
			continue
		}

		p.AddIPAddr(ra)
		retMap[host] = 0
	}

	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		retMap[addr.String()] = rtt
	}
	p.OnIdle = func() {
		wg.Done()
	}
	err := p.Run()
	if err != nil {
		return nil, err
	}

	wg.Wait()
	for i, host := range hosts {
		rtt := retMap[host]
		if rtt == 0 {
			ret[i] = 0
		} else {
			ret[i] = 1
		}
	}

	return ret, nil
}
