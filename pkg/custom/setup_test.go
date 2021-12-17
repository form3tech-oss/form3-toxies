package custom

import (
	toxiclient "github.com/Shopify/toxiproxy/v2/client"
	"github.com/Shopify/toxiproxy/v2/toxics"
	"testing"
	"time"

	"github.com/Shopify/toxiproxy/v2"
)

func TestMain(m *testing.M) {
	toxics.Register("psql", new(PsqlToxic))

	server := toxiproxy.NewServer()

	go server.Listen("localhost", "8474")
	time.Sleep(1 * time.Second)

	client := toxiclient.NewClient("localhost:8474")

	proxy, err := client.CreateProxy("postgres", "localhost:4321", "localhost:5432")
	if err != nil {
		panic(err)
	}

	proxy.AddToxic("psql", "psql", "upstream", 100, nil)

	m.Run()

}
