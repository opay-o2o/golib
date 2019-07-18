package net2

import "net"

func GetLocalIp() (ip string, err error) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ip = ipnet.IP.String()
			return
		}
	}

	return
}
