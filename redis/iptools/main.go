package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	checkHostname()
	checkIPAddress()
}

func checkIPAddress() {
	//checks the IP address against the live redis cluster
	// to prevent the deleation of the production redis db

	// Get live redis cache ip checkIPAddress
	ips, err := net.LookupIP("live-pfcache.prod.aistemos.com")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get IPs: %v\n", err)
		os.Exit(1)
	}
	for _, ip := range ips {
		fmt.Printf("%s\n", ip.String())
	}

	// Get local IP address from all interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		return
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			fmt.Printf("Oops: %v\n", err)
			return
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// process IP address
			fmt.Println(ip)
		}
	}
}

func checkHostname() {
	host, _ := net.LookupHost("live-pfcache.prod.aistemos.com")
	fmt.Println(host)
	cname, _ := net.LookupCNAME("live-pfcache.prod.aistemos.com")
	fmt.Println(cname)
}
