package http

import (
	"testing"

	"github.com/Shopify/toxiproxy/v2"
	"github.com/Shopify/toxiproxy/v2/toxics"
)

func TestMain(m *testing.M) {
	toxics.Register("http", new(HttpToxic))

	server := toxiproxy.NewServer()

	go server.Listen("localhost", toxiProxyPort)
	m.Run()
}
