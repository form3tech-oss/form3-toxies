package custom

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/Shopify/toxiproxy/v2/stream"
	"github.com/Shopify/toxiproxy/v2/toxics"
	"io"
	"strings"
	"time"
)

type PsqlToxic struct{}

type PostgresMessage interface {
	Read(reader io.Reader) ([]byte, error)
	Next() PostgresMessage
	String() string
}

type StartupMessage struct {
	ProtocolVersion int
	MessageLength   int
	Message         []byte
}

type PayloadMessage struct {
	payloadHeader []byte
	MessageType   rune
	MessageLength int
	Message       []byte
}

func (m *StartupMessage) Read(reader io.Reader) ([]byte, error) {
	startupHeader := make([]byte, 8)

	n, err := reader.Read(startupHeader)
	if err != nil {
		return startupHeader[:n], err
	}

	if n != len(startupHeader) {
		return startupHeader[:n], errors.New("malformed startup header")
	}

	m.MessageLength = int(binary.BigEndian.Uint32(startupHeader[0:4])) - 8
	m.ProtocolVersion = int(binary.BigEndian.Uint32(startupHeader[4:8]))
	m.Message = make([]byte, m.MessageLength)
	n, err = reader.Read(m.Message)

	if n != m.MessageLength {
		return append(startupHeader, m.Message[:n]...), errors.New("malformed startup message")
	}

	return append(startupHeader, m.Message[:n]...), err
}

func (m *StartupMessage) String() string {
	return fmt.Sprintf("protoVersion=%d", m.ProtocolVersion)
}

func (m *StartupMessage) Next() PostgresMessage {
	return &PayloadMessage{
		payloadHeader: make([]byte, 5),
	}
}

func (m *PayloadMessage) Read(reader io.Reader) ([]byte, error) {
	n, err := reader.Read(m.payloadHeader)
	if err != nil {
		return m.payloadHeader[:n], err
	}

	if n != len(m.payloadHeader) {
		return m.payloadHeader[:n], errors.New("malformed payload header")
	}

	m.MessageType = rune(m.payloadHeader[0])
	m.MessageLength = int(binary.BigEndian.Uint32(m.payloadHeader[1:5]) - 4)
	m.Message = make([]byte, m.MessageLength)
	n, err = reader.Read(m.Message)

	if n != m.MessageLength {
		return append(m.payloadHeader, m.Message[:n]...), errors.New("malformed payload message")
	}

	return append(m.payloadHeader, m.Message[:n]...), err
}

func (m *PayloadMessage) String() string {
	return fmt.Sprintf("msgLen=%d, type=%c, cmd=%s", m.MessageLength, m.MessageType, strings.TrimRight(string(m.Message), "\000"))
}

func (m *PayloadMessage) Next() PostgresMessage {
	return m
}

func (t *PsqlToxic) Pipe(stub *toxics.ToxicStub) {
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
		stub.Output <- &stream.StreamChunk{
			Data:      read,
			Timestamp: time.Now(),
		}

		fmt.Println(message.String())
		message = message.Next()
	}

}
