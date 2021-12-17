package custom

import (
	"sync"
	"testing"

	"github.com/Shopify/toxiproxy/v2"
)

func TestMain(m *testing.M) {
	server := toxiproxy.NewServer()
	proxy := &toxiproxy.Proxy{
		Mutex:    sync.Mutex{},
		Name:     "postgres",
		Listen:   "4321",
		Upstream: "localhost:5432",
		Enabled:  true,
		Toxics: &toxiproxy.ToxicCollection{
			Mutex: sync.Mutex{},
		},
	}
	err := server.Collection.Add(proxy, true)
	if err != nil {
		panic(err)
	}

	go server.Listen("localhost", "8474")

	m.Run()
}
