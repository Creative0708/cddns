package main

import (
	"io"
	"net"
	"net/http"
)

type IpProvider interface {
	GetIp(client *http.Client) (net.IP, error)
}

type IpifyIpProvider struct {
}

func (_ IpifyIpProvider) GetIp(client *http.Client) (net.IP, error) {
	res, err := client.Get("https://api.ipify.org")
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return net.ParseIP(string(body)), nil
}
