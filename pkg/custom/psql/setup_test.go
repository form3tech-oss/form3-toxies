package psql

import (
	"github.com/Shopify/toxiproxy/v2/toxics"
	"testing"
	"time"

	"github.com/Shopify/toxiproxy/v2"
)

func TestMain(m *testing.M) {
	toxics.Register("psql", new(PostgresToxic))

	server := toxiproxy.NewServer()

	go server.Listen("localhost", "8474")
	time.Sleep(1 * time.Second)

	m.Run()
}
