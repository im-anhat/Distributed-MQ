package producer

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"github.com/im-anhat/Distributed-MQ/internal/broker"
	"github.com/im-anhat/Distributed-MQ/internal/wire"
)

type Producer struct {
	Port    uint16
	TopicID uint16
}

func (p *Producer) registerWithBroker() error {
	var err error
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", broker.BROKER_PORT))
	if err != nil {
		return err
	}
	defer conn.Close()
	stream_rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	p_reg := wire.ProducerRegisterMessage{}
	p_reg.Port = p.Port
	p_reg.TopicID = p.TopicID
	fmt.Printf("pRegMsg: port=%d, topicID=%d\n", p_reg.Port, p_reg.TopicID)
	err = wire.WriteMessageToStream(stream_rw, &wire.Message{P_REG: &p_reg})
	if err != nil {
		return err
	}

	message, err := wire.ReadMessageFromStream(stream_rw)
	if err != nil {
		return err
	}
	fmt.Printf("Received response from broker: %d\n", *message.R_P_REG)
	return nil
}

func (p *Producer) StartProducerServer() error {
	// Start producer server
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", p.Port))
	if err != nil {
		return err
	}
	fmt.Println("Producer server started.")

	// Register with broker first
	err = p.registerWithBroker()
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

		// Write PCM
		err = wire.WriteMessageToStream(stream_rw, &wire.Message{PCM: []byte(line)})
		if err != nil {
			break
		}

		// Read message from stream
		resp_message, err := wire.ReadMessageFromStream(stream_rw)
		if err != nil {
			break
		}
		fmt.Printf("Received R_PCM from broker: %d", *resp_message.R_PCM)
	}

	err = conn.Close()
	if err != nil {
		return err
	}
	return nil
}
