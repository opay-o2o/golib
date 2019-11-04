package net2

import (
	"net"
	"strings"
)

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

func GetInterfaceIp(name string) (ip string, err error) {
	inter, err := net.InterfaceByName(name)

	if err != nil {
		return
	}

	addrs, err := inter.Addrs()

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

func InIpList(ip string, ips []string) bool {
	items := strings.Split(ip, ".")

	if len(items) != 4 {
		return false
	}

	ipC := strings.Join(items[:3], ".") + ".*"
	ipB := strings.Join(items[:2], ".") + ".*.*"

	for _, v := range ips {
		if v == ip || v == ipC || v == ipB {
			return true
		}
	}

	return false
}
