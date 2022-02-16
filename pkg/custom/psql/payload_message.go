package psql

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type PayloadMessage struct {
	payloadHeader []byte
	MessageType   rune
	MessageLength int
	Message       []byte
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

func (m *PayloadMessage) HasStatement(statement string) bool {
	match, err := regexp.MatchString(statement, string(m.Message))
	if err != nil {
		fmt.Println(err)
	}
	return match
}

func (m *PayloadMessage) Next() PostgresMessage {
	return m
}
