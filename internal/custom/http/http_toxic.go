package http

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Shopify/toxiproxy/v2/toxics"
)

type Condition struct {
	FailOn       int
	RecoverAfter int
	Method       string
}

type HttpToxic struct {
	count     map[string]int
	Condition map[string]Condition
}

func (t *HttpToxic) Pipe(stub *toxics.ToxicStub) {
	for {
		select {
		case <-stub.Interrupt:
			return
		case c := <-stub.Input:

			if c == nil {
				stub.Close()
				return
			}
			fmt.Println("received data in http toxic ", string(c.Data))
			method, path := getHTTPMethodAndPath(c.Data)
			fmt.Printf("method is %s path is %s \n", method, path)
			for pathRegex, condition := range t.Condition {
				match, err := regexp.MatchString(pathRegex, path)
				if err != nil {
					fmt.Printf("Error %s when matching path with regex", err)
					continue
				}
				if match {
					if t.count == nil {
						t.count = make(map[string]int)
					}
					count := t.count[path]
					count++
					t.count[path] = count
					if method == condition.Method && count >= condition.FailOn && count <= condition.RecoverAfter {
						fmt.Println("condition matched in http proxy failing.")
						stub.Close()
						return
					}
				}
			}
			stub.Output <- c
		}
	}
}

func getHTTPMethodAndPath(data []byte) (string, string) {
	var method strings.Builder
	var path strings.Builder
	methodFound := false
	pathFound := false
	for _, v := range string(data) {
		ch := string(v)
		if !methodFound {
			if ch == " " {
				methodFound = true
				continue
			}
			method.WriteString(ch)
		}
		if methodFound && !pathFound {
			if ch == " " {
				pathFound = true
				break
			}
			path.WriteString(ch)
		}
	}
	return method.String(), path.String()
}
