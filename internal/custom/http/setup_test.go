package http

import (
	"testing"
	"time"

	"github.com/Shopify/toxiproxy/v2"
	"github.com/Shopify/toxiproxy/v2/toxics"
)

func TestMain(m *testing.M) {
	toxics.Register("http", new(HttpToxic))

	server := toxiproxy.NewServer()

	go server.Listen("localhost", toxiProxyPort)
	time.Sleep(2 * time.Second)
	m.Run()
}
