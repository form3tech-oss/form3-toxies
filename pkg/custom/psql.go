package custom

import (
	"encoding/binary"
	"fmt"
	"github.com/Shopify/toxiproxy/v2/stream"
	"github.com/Shopify/toxiproxy/v2/toxics"
	"io"
	"strings"
)

type PsqlToxic struct{}

func (t *PsqlToxic) Pipe(stub *toxics.ToxicStub) {
	//buf := make([]byte, 32*1024)
	writer := stream.NewChanWriter(stub.Output)
	reader := stream.NewChanReader(stub.Input)
	reader.SetInterrupt(stub.Interrupt)

	readStartupHeader := false

	startupHeader := make([]byte, 8)
	payloadHeader := make([]byte, 5)
	upstream := make([]byte, 0)

	for {
		length := int32(0)

		if !readStartupHeader {
			_, _ = reader.Read(startupHeader)
			readStartupHeader = true

			protoVersion := int32(0)

			/*binary.Read(bytes.NewBuffer(startupHeader[0:4]), binary.BigEndian, &length)
			binary.Read(bytes.NewBuffer(startupHeader[4:8]), binary.BigEndian, &protoVersion)

			length = length - int32(n)*/
			length = int32(binary.BigEndian.Uint32(startupHeader[0:4]) - 8)
			protoVersion = int32(binary.BigEndian.Uint32(startupHeader[4:8]))

			fmt.Printf("protoVersion=%d\n", protoVersion)
			upstream = startupHeader
		} else {
			_, _ = reader.Read(payloadHeader)

			msgType := rune(payloadHeader[0])
			fmt.Printf("type=%c\n", msgType)

			length = int32(binary.BigEndian.Uint32(payloadHeader[1:5]) - 4)
			upstream = payloadHeader
		}

		msgbuf := make([]byte, length)
		n, err := reader.Read(msgbuf)

		fmt.Printf("len=%d, msgLen=%d, cmd=%s\n", length, n, strings.TrimRight(string(msgbuf[:n]), "\000"))

		if err == stream.ErrInterrupted {
			writer.Write(append(upstream, msgbuf[:n]...))
			return
		} else if err == io.EOF {
			stub.Close()
			return
		}
		writer.Write(append(upstream, msgbuf[:n]...))
	}

}
