package http_mock

import (
	"fmt"
	"github.com/Shopify/toxiproxy/v2/toxics"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
)

type FailureType string
type HTTPMockToxic struct {
	FailureType  FailureType
	Method       string
	Path         string
	FailOn       int
	RecoverAfter int

	reqCount uint32
	paths    sync.Map
}

func (t *HTTPMockToxic) Pipe(stub *toxics.ToxicStub) {
	methodAndLocation := strings.TrimSpace(fmt.Sprintf("%s %s", t.Method, t.Path))
	if methodAndLocation == "" {
		reqCount := int(atomic.AddUint32(&t.reqCount, 1))
		if (t.FailOn > 0 && reqCount >= t.FailOn) && (t.RecoverAfter == 0 || reqCount <= t.RecoverAfter) {
			stub.Close()
			return
		}
	}

	if t.Method == "" && t.Path != "" {
		methodAndLocation = fmt.Sprintf(".*? %s", t.Path)
	}

	firstChunk := true

	for {
		select {
		case <-stub.Interrupt:
			return
		case c := <-stub.Input:
			if c == nil {
				stub.Output <- c
				continue
			}
			chunk := string(c.Data)
			_ = chunk
			if methodAndLocation != "" && firstChunk {
				firstChunk = false
				ok, err := regexp.Match(fmt.Sprintf("^%s HTTP", methodAndLocation), c.Data)
				if err != nil {
					stub.Close()
					return
				}
				if ok {
					v, ok := t.paths.Load(t.Path)
					if !ok {
						v = 0
					}
					reqCount := v.(int) + 1
					t.paths.Store(t.Path, reqCount)
					if (t.FailOn > 0 && reqCount >= t.FailOn) && (t.RecoverAfter == 0 || reqCount <= t.RecoverAfter) {
						stub.Close()
						return
					}
				}
			}
			if c == nil {
				stub.Close()
				return
			}
			stub.Output <- c
		}

	}

}
