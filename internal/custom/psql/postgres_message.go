package psql

import (
	"io"
)

type PostgresMessage interface {
	Read(reader io.Reader) ([]byte, error)
	Next() PostgresMessage
	String() string
	HasStatement(statement string) bool
}
