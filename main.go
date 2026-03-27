package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	// fmt.Println(os.Args)
	// if os.Args[1] == "server" {
	// 	// startServer()
	// 	spawnServer()
	// } else {
	// 	clientConnect(os.Args[2])
	// }

	fmt.Println("Args: ", os.Args)
	if os.Args[1] == "server" {
		broker := Broker{}
		err := broker.startBrokerServer()
		if err != nil {
			fmt.Println("Error starting broker server:", err)
		}
	} else {
		clientConnectTCPAndEcho(10000)
	}
}

func writeEchoToStream(stream_rw *bufio.ReadWriter, data string) error {
	var err error
	err = stream_rw.WriteByte(byte(len(data) + 1)) // Write length first, +1 for message type
	if err != nil {
		return err
	}

	err = stream_rw.WriteByte(byte(ECHO)) // Write message type
	if err != nil {
		return err
	}

	_, err = stream_rw.WriteString(data) // Write actual data
	if err != nil {
		return err
	}
	stream_rw.Flush()
	return nil
}

func clientConnectTCPAndEcho(port int) {
	conn, _ := net.Dial("tcp", fmt.Sprintf(":%d", port))
	rd := bufio.NewReader(os.Stdin)
	stream_rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	line, err := rd.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			fmt.Println("End of input")
		} else {
			// panic(err)
		}
	}

	fmt.Println("Connected to server, sending: ", line)
	err = writeEchoToStream(stream_rw, strings.Trim(line, "\n"))
	if err != nil {
		fmt.Println("Error writing to stream:", err)
	}

	fmt.Println("Read back from broker...")
	header, err := stream_rw.ReadByte()
	if header == 0 || err != nil {
		return
	}
	data, _ := stream_rw.Peek(int(header))
	fmt.Printf("Data from server: %s\n", data)
	stream_rw.Discard(int(header))
	conn.Close()
}
