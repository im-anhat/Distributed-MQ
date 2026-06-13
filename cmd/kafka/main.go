package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	brokerpkg "github.com/im-anhat/Distributed-MQ/internal/broker"
	consumerpkg "github.com/im-anhat/Distributed-MQ/internal/consumer"
	producerpkg "github.com/im-anhat/Distributed-MQ/internal/producer"
	"github.com/im-anhat/Distributed-MQ/internal/wire"
)

func main() {
	fmt.Println("Args: ", os.Args)
	if os.Args[1] == "broker" {
		broker := brokerpkg.Broker{}
		broker.Init()
		err := broker.StartBrokerServer()
		if err != nil {
			fmt.Println("Error starting broker server:", err)
		}
	} else if os.Args[1] == "producer" {
		fmt.Println("Trying to start producer processes")
		port, err := strconv.ParseInt(os.Args[2], 10, 16)
		if err != nil {
			panic(err)
		}
		topicID, err := strconv.ParseInt(os.Args[3], 10, 16)
		if err != nil {
			panic(err)
		}
		producer := producerpkg.Producer{}
		producer.Port = uint16(port)
		producer.TopicID = uint16(topicID)
		fmt.Printf("producer: port=%d, topicID=%d\n", producer.Port, producer.TopicID)
		producer.StartProducerServer()
	} else if os.Args[1] == "consumer" {
		fmt.Println("Trying to start producer processes")
		port, err := strconv.ParseInt(os.Args[2], 10, 16)
		if err != nil {
			panic(err)
		}
		topicID, err := strconv.ParseInt(os.Args[3], 10, 16)
		if err != nil {
			panic(err)
		}
		groupID, err := strconv.ParseInt(os.Args[4], 10, 16)
		if err != nil {
			panic(err)
		}
		consumer := &consumerpkg.Consumer{}
		consumer.Port = uint16(port)
		consumer.TopicID = uint16(topicID)
		consumer.GroupID = uint16(groupID)
		fmt.Printf("consumer: port=%d, topicID=%d, groupID=%d\n", consumer.Port, consumer.TopicID, consumer.GroupID)
		consumer.StartConsumerServer()
	} else {
		panic("Invalid argument")
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
	err = wire.WriteMessageToStream(stream_rw, &wire.Message{ECHO: &message})
	if err != nil {
		panic(err)
	}

	// Read from stream
	resp_message, err := wire.ReadMessageFromStream(stream_rw)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Received message from server: %s", *resp_message.R_ECHO)
	conn.Close()
}
