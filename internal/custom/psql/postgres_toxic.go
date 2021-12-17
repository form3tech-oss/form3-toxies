package psql

import (
	"encoding/binary"
	"io"
	"sync/atomic"
	"time"

	"github.com/Shopify/toxiproxy/v2/stream"
	"github.com/Shopify/toxiproxy/v2/toxics"
)

type FailureType string

const FailureTypeSyntaxError = "SyntaxError"
const FailureTypeConnectionFailure = "ConnectionFailure"

type PostgresToxic struct {
	FailureType  FailureType
	SearchText   string
	FailOn       int
	RecoverAfter int
	matched      uint32
}

func BadMessage() []byte {
	result := make([]byte, 7)
	result[0] = 'Q'
	binary.BigEndian.PutUint32(result[1:], uint32(len(result)-1))
	result[5] = 0
	result[6] = 0
	return result
}

func (t *PostgresToxic) Pipe(stub *toxics.ToxicStub) {
	reader := stream.NewChanReader(stub.Input)
	reader.SetInterrupt(stub.Interrupt)

	var message PostgresMessage = &StartupMessage{}
	for {
		read, err := message.Read(reader)

		if err == stream.ErrInterrupted {
			stub.Output <- &stream.StreamChunk{
				Data:      read,
				Timestamp: time.Now(),
			}
			return
		} else if err == io.EOF {
			stub.Close()
			return
		}

		if message.HasStatement(t.SearchText) {
			matched := int(atomic.AddUint32(&t.matched, 1))
			if (t.FailOn > 0 && matched >= t.FailOn) && (t.RecoverAfter == 0 || matched <= t.RecoverAfter) {
				if t.FailureType == FailureTypeConnectionFailure {
					stub.Close()
					return
				}
				read = BadMessage()
			}
		}

		stub.Output <- &stream.StreamChunk{
			Data:      read,
			Timestamp: time.Now(),
		}

		message = message.Next()
	}

}
