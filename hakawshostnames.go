package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type AWSIPs struct {
	SyncToken  string        `json:"syncToken"`
	CreateDate string        `json:"createDate"`
	Prefixes   []interface{} `json:"prefixes"`
}

func main() {
	var response AWSIPs

	client := &http.Client{Timeout: 10 * time.Second}
	r, err := client.Get("https://ip-ranges.amazonaws.com/ip-ranges.json")
	if err != nil {
		log.Println("Error fetching ip ranges from AWS:", err)
		os.Exit(1)
	}
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		log.Println("Error decoding response:", err)
		os.Exit(1)
	}

	w := bufio.NewWriter(os.Stdout)
	for _, prefixInterface := range response.Prefixes {
		prefix := prefixInterface.(map[string]interface{})

		// if !ok {
		// 	fmt.Println("not ok")
		// 	continue
		// }
		ips, err := ExpandCIDR(prefix["ip_prefix"].(string))
		if err != nil {
			log.Println("Error expanding CIDR:", err)
		}
		for _, ip := range ips {
			ipString := strings.ReplaceAll(ip, ".", "-")
			//example: ec2-35-180-0-1.eu-west-3.compute.amazonaws.com
			fmt.Fprintf(w, "ec2-%s.%s.compute.amazonaws.com\n", ipString, prefix["region"].(string))
		}
	}
}

// Expands CIDR notation to a slice of IP address strings
func ExpandCIDR(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// remove network address and broadcast address
	lenIPs := len(ips)
	switch {
	case lenIPs < 2:
		return ips, nil

	default:
		return ips[1 : len(ips)-1], nil
	}
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
