package http_mock

import (
	"context"
	"fmt"
	toxiclient "github.com/Shopify/toxiproxy/v2/client"
	"net/http"
	"strings"
	"testing"
)

const toxicName = "http_mock"
const proxyName = "httpMockProxy"

type failureType string

type options struct {
	Path         string
	FailureType  failureType
	FailOn       int
	RecoverAfter int
}

type httpMockStage struct {
	incomingRequest  *http.Request
	receivedResponse []*http.Response
	receivedError    []error

	mockServerPort int
	proxyPort      int
	t              *testing.T
}

func HTTPMockTest(t *testing.T) (*httpMockStage, *httpMockStage, *httpMockStage) {
	proxyPort, err := getFreePort()
	if err != nil {
		t.Fatal(err)
	}
	mockServerPort, err := getFreePort()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		proxy, err := toxiclient.NewClient(fmt.Sprintf("localhost:%d", toxiProxyPort)).Proxy(proxyName)
		if err == nil && proxy != nil {
			proxy.Delete()
		}
	})

	s := &httpMockStage{
		t:                t,
		mockServerPort:   mockServerPort,
		proxyPort:        proxyPort,
		receivedResponse: make([]*http.Response, 0),
		receivedError:    make([]error, 0),
	}
	return s, s, s
}

func (s *httpMockStage) and() *httpMockStage {
	return s
}

func (s *httpMockStage) an_get_request() {
	s.an_get_request_with_path("/")
}
func (s *httpMockStage) an_get_request_with_path(path string) {
	var err error
	if strings.Index(path, "/") == 0 {
		path = path[1:]
	}
	s.incomingRequest, err = http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/%s", s.proxyPort, path), nil)
	if err != nil {
		s.t.Fatal("failed creating request")
	}
}

func (s *httpMockStage) request_is_sent() {
	s.request_is_sent_n_times(1)
}

func (s *httpMockStage) request_is_sent_n_times(n int) {
	client := &http.Client{}
	for i := 0; i < n; i++ {
		resp, err := client.Do(s.incomingRequest)
		s.receivedResponse = append(s.receivedResponse, resp)
		s.receivedError = append(s.receivedError, err)
	}
}

func (s *httpMockStage) an_server_accepting_any_requests() *httpMockStage {
	mux := http.NewServeMux()
	mux.HandleFunc("*", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
	})

	go func() {
		server := &http.Server{Addr: fmt.Sprintf(":%d", s.mockServerPort), Handler: mux}

		if err := server.ListenAndServe(); err != nil {
			s.t.Fatalf("failed serving up: %+v", err)
		}

		s.t.Cleanup(func() {
			if err := server.Shutdown(context.Background()); err != nil {
				s.t.Logf("failed shutting down server: %+v", err)
			}
		})
	}()

	return s
}
func (s *httpMockStage) an_http_mock_toxic() *httpMockStage {
	return s.an_http_mock_toxic_with_options(options{})
}

func (s *httpMockStage) an_http_mock_toxic_with_options(o options) *httpMockStage {
	client := toxiclient.NewClient(fmt.Sprintf("localhost:%d", toxiProxyPort))

	proxy, err := client.CreateProxy(proxyName,
		fmt.Sprintf("localhost:%d", s.proxyPort),
		fmt.Sprintf("localhost:%d", s.mockServerPort))

	if err != nil {
		s.t.Fatal(err)
	}

	s.t.Cleanup(func() {
		_ = proxy.RemoveToxic(toxicName)
	})

	_, err = proxy.AddToxic(toxicName, "http_mock", "upstream", 100, map[string]interface{}{
		"FailOn":       o.FailOn,
		"Path":         o.Path,
		"RecoverAfter": o.RecoverAfter,
		"FailureType":  o.FailureType,
	})

	if err != nil {
		s.t.Fatal(err)
	}

	return s
}

func (s *httpMockStage) request_was_successful() {
	if len(s.receivedError) == 0 {
		s.t.Fatalf("request failed: %+v", s.receivedError)
	}
	for _, err := range s.receivedError {
		if err != nil {
			s.t.Fatalf("request failed: %+v", err)
		}
	}
}

func (s *httpMockStage) requests_were_successful() {
	s.request_was_successful()
}

func (s *httpMockStage) request_started_failing_on_attempt(attempt int) {
	if len(s.receivedResponse) < attempt || len(s.receivedError) < attempt {
		s.t.Fatal("not enough requests were made to perform this check")
	}
	hasFailures := false
	expectedIndex := attempt - 1
	for i := range s.receivedResponse {
		if i >= expectedIndex {
			if s.receivedError[i] == nil {
				s.t.Fatalf("request attempt #%d did not fail", i+1)
			}
			continue
		}
		if s.receivedError[i] != nil {
			s.t.Logf("request attempt #%d expected to succeed but failed: %+v", i+1, s.receivedError[i])
			hasFailures = true
		}
	}
	if hasFailures {
		s.t.Fail()
	}
}

func (s *httpMockStage) request_failed_on_attempt(attempt int) {
	if len(s.receivedResponse) < attempt || len(s.receivedError) < attempt {
		s.t.Fatal("not enough requests were made to perform this check")
	}
	hasFailures := false
	expectedIndex := attempt - 1
	for i := range s.receivedResponse {
		if i == expectedIndex {
			if s.receivedError[i] == nil {
				s.t.Fatalf("request attempt #%d did not fail", i+1)
			}
			continue
		}
		if s.receivedError[i] != nil {
			s.t.Logf("request attempt #%d expected to succeed but failed: %+v", i+1, s.receivedError[i])
			hasFailures = true
		}
	}
	if hasFailures {
		s.t.Fail()
	}
}
