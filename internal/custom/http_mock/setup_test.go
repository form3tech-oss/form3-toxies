package http_mock

import (
	"github.com/Shopify/toxiproxy/v2"
	"github.com/Shopify/toxiproxy/v2/toxics"
	"net"
	"strconv"
	"testing"
)

var toxiProxyPort int

func TestMain(m *testing.M) {
	toxics.Register("http_mock", new(HTTPMockToxic))

	server := toxiproxy.NewServer()
	proxyPort, err := getFreePort()
	if err != nil {
		panic(err)
	}

	go server.Listen("localhost", strconv.Itoa(proxyPort))

	toxiProxyPort = proxyPort
	m.Run()
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
