package consumer

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/im-anhat/Distributed-MQ/internal/config"
	"github.com/im-anhat/Distributed-MQ/internal/wire"
)

type Consumer struct {
	Port    uint16
	TopicID uint16
	GroupID uint16
}

func (c *Consumer) registerWithBroker() error {
	var err error
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", config.BrokerPort))
	if err != nil {
		return err
	}
	defer conn.Close()
	stream_rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	c_reg := wire.ConsumerRegisterMessage{}
	c_reg.Port = c.Port
	c_reg.TopicID = c.TopicID
	c_reg.GroupID = c.GroupID
	fmt.Printf("pRegMsg: port=%d, topicID=%d, groupID=%d\n", c_reg.Port, c_reg.TopicID, c_reg.GroupID)
	err = wire.WriteMessageToStream(stream_rw, &wire.Message{C_REG: &c_reg})
	if err != nil {
		return err
	}

	message, err := wire.ReadMessageFromStream(stream_rw)
	if err != nil {
		return err
	}
	fmt.Printf("Received response from broker: %d\n", *message.R_C_REG)
	return nil
}

func (p *Consumer) StartConsumerServer() error {
	// Start Consumer server
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", p.Port))
	if err != nil {
		return err
	}
	fmt.Println("Consumer server started.")

	// Register with broker first
	err = p.registerWithBroker()
	if err != nil {
		return err
	}

	conn, _ := ln.Accept() // Block
	fmt.Println("Consumer server accepted connection.")

	// Read/Write buffer
	stream_rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	for {
		// Read message to consume
		message, err := wire.ReadMessageFromStream(stream_rw)
		if err != nil {
			break
		}
		fmt.Printf("Receive PCM from broker: %s\n", message.PCM)
		time.Sleep(1 * time.Second)
		// Write R_PCM
		var resp byte = 1
		err = wire.WriteMessageToStream(stream_rw, &wire.Message{
			R_PCM: &resp,
		})
		if err != nil {
			break
		}
	}

	err = conn.Close()
	if err != nil {
		return err
	}
	return nil
}
