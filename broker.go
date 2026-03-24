package main

import (
	"bufio"
	"fmt"
	"net"
)

const BROKER_PORT = 10000

func readFromStream(stream_rw *bufio.ReadWriter) (*string, error) {
	var err error

	header, err := stream_rw.ReadByte() // Block
	if err != nil {
		return nil, err
	}

	data, err := stream_rw.Peek(int(header)) // Block
	if err != nil {
		return nil, err
	}
	fmt.Printf("Data from client: %s\n", data)
	dataStr := string(data)
	return &dataStr, nil
}

func writeToStream(stream_rw *bufio.ReadWriter, data string) error {
	var err error

	err = stream_rw.WriteByte(byte(len(data)))
	if err != nil {
		return err
	}
	_, err = stream_rw.WriteString(data)
	if err != nil {
		return err
	}
	stream_rw.Flush()
	return nil
}

type Broker struct {
}

func (b *Broker) startBrokerServer() error {
	ln, _ := net.Listen("tcp", fmt.Sprintf(":%d", BROKER_PORT))
	for {
		conn, _ := ln.Accept() // Block until can
		stream_rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		data, err := readFromStream(stream_rw)
		if err != nil {
			return err
		}

		err = writeToStream(stream_rw, *data)
		if err != nil {
			return err
		}
		err = conn.Close()
		if err != nil {
			return err
		}
	}
}
