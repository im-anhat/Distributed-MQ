package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("Args: ", os.Args)
	if os.Args[1] == "broker" {
		broker := Broker{}
		err := broker.startBrokerServer()
		if err != nil {
			fmt.Println("Error starting broker server:", err)
		}
	} else if os.Args[1] == "producer" {
		producer := Producer{}
		port, err := strconv.Atoi(os.Args[2])
		if err != nil {
			panic(err)
		}
		err = producer.startProducerServer(int16(port))
		if err != nil {
			panic(err)
		}
	} else {
		clientConnectTCPAndEcho(10000)
	}
}

func clientConnectTCPAndEcho(port int) {
	conn, _ := net.Dial("tcp", fmt.Sprintf(":%d", port))
	rd := bufio.NewReader(os.Stdin)
	stream_rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	line, err := rd.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return
		} else {
			// panic(err)
		}
	}

	fmt.Println("Connected to server, sending: ", line)

	// Write to stream
	message := strings.Trim(line, "\n")
	fmt.Printf("Sending message: %s", message)
	err = writeMessageToStream(stream_rw, &Message{ECHO: &message})
	if err != nil {
		panic(err)
	}

	// Read from stream
	resp_message, err := readMessageFromStream(stream_rw)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Received message from server: %s", *resp_message.R_ECHO)
	conn.Close()
}
