package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	toxiclient "github.com/Shopify/toxiproxy/v2/client"
)

const (
	proxyName     = "http"
	toxicName     = "mockAPI"
	toxiProxyPort = "29000"
)

type httpTestStage struct {
	t              *testing.T
	httpServerPort string
	proxyPort      string
	httpServer     *http.Server
	toxyProxy      *toxiclient.Proxy
	receivedData   string
}

func httpTest(t *testing.T) (*httpTestStage, *httpTestStage, *httpTestStage) {
	stage := &httpTestStage{
		t:              t,
		httpServerPort: "40000",
		proxyPort:      "30000",
	}
	return stage, stage, stage
}

func (s *httpTestStage) and() *httpTestStage {
	return s
}

func (s *httpTestStage) a_http_server() *httpTestStage {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("successfully served slash")
		fmt.Fprintf(w, "Hello World!")
	})
	httpServer := &http.Server{
		Addr:    net.JoinHostPort("localhost", s.httpServerPort),
		Handler: mux,
	}
	s.httpServer = httpServer
	fmt.Printf("Starting server at port %s\n", s.httpServerPort)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatal("error starting http server", err)
			}
			log.Print("server closed gracefully")
		}
	}()
	time.Sleep(2 * time.Second)
	s.t.Cleanup(func() {
		fmt.Println("proxy and http server shutdown")
		s.httpServer.Shutdown(context.TODO())
		s.toxyProxy.Delete()
	})
	return s
}
func (s *httpTestStage) a_http_toxic(proxyOption map[string]Condition) *httpTestStage {
	client := toxiclient.NewClient(net.JoinHostPort("localhost", toxiProxyPort))
	proxy, err := client.CreateProxy(proxyName,
		net.JoinHostPort("localhost", s.proxyPort),
		net.JoinHostPort("localhost", s.httpServerPort))
	s.toxyProxy = proxy
	if err != nil {
		s.t.Fatal(err)
	}
	toxyAttributes := map[string]interface{}{
		"condition": proxyOption,
	}

	_, err = proxy.AddToxic(toxicName, "http", "upstream", 100, toxyAttributes)

	if err != nil {
		s.t.Fatal(err)
	}
	return s
}

func (s *httpTestStage) a_http_call_succeeds(path string, httpMethod string) *httpTestStage {
	fmt.Println("calling http success")
	httpClient := http.DefaultClient
	httpClient.Timeout = 5 * time.Second
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	url := net.JoinHostPort("localhost", s.proxyPort)
	url = "http://" + url + path
	httpRequest, err := http.NewRequestWithContext(ctx, httpMethod, url, nil)
	if err != nil {
		s.t.Log(err)
		s.t.FailNow()
	}
	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		s.t.Log(err)
		s.t.FailNow()
	}
	if resp.StatusCode != http.StatusOK {
		s.t.Logf("Expected 200 status code but got %d", resp.StatusCode)
		s.t.FailNow()
	}
	receivedResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.t.FailNow()
	}
	defer resp.Body.Close()

	s.receivedData = string(receivedResp)
	fmt.Println("received data in handler", s.receivedData)
	return s
}

func (s *httpTestStage) a_http_call_fails(path string, httpMethod string) *httpTestStage {
	fmt.Println("calling http failure")
	httpClient := http.DefaultClient
	httpClient.Timeout = 5 * time.Second
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	url := net.JoinHostPort("localhost", s.proxyPort)
	url = "http://" + url + path
	httpRequest, err := http.NewRequestWithContext(ctx, httpMethod, url, nil)
	if err != nil {
		s.t.Log(err)
		s.t.FailNow()
	}
	resp, err := httpClient.Do(httpRequest)
	if err == nil {
		return s
	}
	if resp.StatusCode != http.StatusOK {
		s.t.Logf("Expected 200 status code but got %d", resp.StatusCode)
		s.t.FailNow()
	}
	receivedResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.t.FailNow()
	}
	defer resp.Body.Close()
	s.receivedData = string(receivedResp)
	fmt.Println("received data in failure handler", s.receivedData)
	return s
}
