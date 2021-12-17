package custom

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/Shopify/toxiproxy/v2/stream"
	"github.com/Shopify/toxiproxy/v2/toxics"
	"io"
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
		v := make([]byte, 4)
		binary.BigEndian.PutUint32(v, 196608)

		sample := make([]byte, 1000)
		reader.Read(sample)

		length := int32(0)

		if !readStartupHeader {
			_, _ = reader.Read(startupHeader)
			readStartupHeader = true

			length = int32(binary.BigEndian.Uint32(startupHeader[0:4]))
			protoVersion := binary.BigEndian.Uint32(startupHeader[4:8])

			//binary.Read(bytes.NewBuffer(startupHeader[0:4]), binary.LittleEndian, &length)
			//binary.Read(bytes.NewBuffer(startupHeader[4:8]), binary.LittleEndian, &protoVersion)

			length = length - 4
			//length = int(binary.BigEndian.Uint32(startupHeader[0:4]) - 4)

			fmt.Printf("protoVersion=%d\n", protoVersion)
			upstream = startupHeader
		} else {
			_, _ = reader.Read(payloadHeader)

			msgType := payloadHeader[0]
			fmt.Printf("type=%d\n", msgType)

			binary.Read(bytes.NewBuffer(startupHeader[0:4]), binary.LittleEndian, &length)

			upstream = payloadHeader
		}

		msgbuf := make([]byte, length)
		_, err := reader.Read(msgbuf)

		fmt.Printf("cmd=%s\n", string(msgbuf))

		if err == stream.ErrInterrupted {
			writer.Write(append(upstream, msgbuf...))
			return
		} else if err == io.EOF {
			stub.Close()
			return
		}
		writer.Write(append(upstream, msgbuf...))
	}

}
