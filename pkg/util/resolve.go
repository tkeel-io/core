package util

import (
	"log"
	"net"
)

func ResolveAddr() string {
	var (
		err   error
		hosts []string
		addrs []net.Addr
	)
	if addrs, err = net.InterfaceAddrs(); err != nil {
		log.Fatal(err)
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				hosts = append(hosts, ipnet.IP.String())
			}
		}
	}

	if len(hosts) == 0 {
		log.Fatal("no avalible address")
	}

	// select address from hosts.
	return hosts[0]
}
