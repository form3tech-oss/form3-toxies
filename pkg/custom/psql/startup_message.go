package psql

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type StartupMessage struct {
	ProtocolVersion int
	MessageLength   int
	Message         []byte
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

func (m *StartupMessage) HasStatement(_ string) bool {
	return false
}

func (m *StartupMessage) Next() PostgresMessage {
	return &PayloadMessage{
		payloadHeader: make([]byte, 5),
	}
}
