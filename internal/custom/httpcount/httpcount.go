package httpcount

import (
	"fmt"

	"github.com/Shopify/toxiproxy/v2/toxics"
)

type Condition struct {
	Path         string
	FailAfter    int
	SucceedAfter int
}
type HttpCountToxic struct {
	Condition Condition
	count     int
}

func (t *HttpCountToxic) Pipe(stub *toxics.ToxicStub) {
	for {
		select {
		case <-stub.Interrupt:
			return
		case c := <-stub.Input:
			if c == nil {
				stub.Close()
				return
			}
			t.count = t.count + 1
			if t.count > t.Condition.FailAfter && t.count <= t.Condition.SucceedAfter {
				fmt.Println("condition matched in http proxy failing.")
				stub.Close()
				return
			}

			fmt.Println("received data in http toxic ", string(c.Data))
			stub.Output <- c
		}
	}
}
