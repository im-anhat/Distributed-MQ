package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

type Producer struct {
}

func (p *Producer) registerWithBroker(port int16) error {
	var err error
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", BROKER_PORT))
	if err != nil {
		return err
	}
	defer conn.Close()

	stream_rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	port_str := fmt.Sprintf("%d", port)
	err = writeMessageToStream(stream_rw, &Message{P_REG: &port_str})
	if err != nil {
		return err
	}

	message, err := readMessageFromStream(stream_rw)
	if err != nil {
		return err
	}
	fmt.Printf("Received response from broker: %d\n", *message.R_P_REG)
	return nil
}

func (p *Producer) startProducerServer(port int16) error {
	// Start producer server
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	fmt.Println("Producer server started.")

	// Register with broker first
	err = p.registerWithBroker(port)
	if err != nil {
		return err
	}

	conn, _ := ln.Accept() // Block
	fmt.Println("Producer server accepted connection.")

	// Read/Write buffer
	rd := bufio.NewReader(os.Stdin)
	stream_rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	for {
		// Read from stdin
		line, err := rd.ReadString('\n')
		if err != nil {
			break
		}

		// Write message to stream
		err = writeMessageToStream(stream_rw, &Message{ECHO: &line})
		if err != nil {
			break
		}

		// Read message from stream
		resp_message, err := readMessageFromStream(stream_rw)
		if err != nil {
			break
		}
		fmt.Printf("Received message from broker: %s", *resp_message.R_ECHO)
	}

	err = conn.Close()
	if err != nil {
		return err
	}
	return nil
}
