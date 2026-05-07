package main

import (
	"bufio"
	"fmt"
	"net"
)

const BROKER_PORT = 10000

type Broker struct {
	topics []Topic
}

func (b *Broker) init() {
	b.topics = make([]Topic, 0)
}

// bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
func (b *Broker) startBrokerServer() error {
	ln, _ := net.Listen("tcp", fmt.Sprintf(":%d", BROKER_PORT))
	fmt.Println("Server started...")
	for {
		conn, _ := ln.Accept() // Block
		stream_rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

		message, err := readMessageFromStream(stream_rw)
		if err == nil && message != nil {
			resp, err := b.processBrokerMessage(message)
			if err != nil {
				return err
			}

			// Write it back
			err = writeMessageToStream(stream_rw, resp)
			if err != nil {
				return err
			}
		}

		err = conn.Close()
		if err != nil {
			return err
		}
	}
}

func (b *Broker) processBrokerMessage(message *Message) (*Message, error) {
	var err error
	var resp *Message

	if message.ECHO != nil {
		resp, err = b.processEchoMessage(message.ECHO)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
	if message.P_REG != nil {
		resp, err = b.processProducerRegisterMessage(message.P_REG)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	return resp, err
}

func (b *Broker) processProducerPCM(pcm []byte, topicIdx uint16) (*Message, error) {
	b.topics[topicIdx].mq.push(pcm)
	b.topics[topicIdx].mq.debug()
	one := byte(1)
	return &Message{R_PCM: &one}, nil
}

func (b *Broker) processEchoMessage(echo_message *string) (*Message, error) {
	fmt.Printf("Received Echo message: %s!", *echo_message)
	resp_echo := fmt.Sprintf("I have received your message: %s", *echo_message)
	return &Message{R_ECHO: &resp_echo}, nil
}

func (b *Broker) processProducerRegisterMessage(reg_message *ProducerRegisterMessage) (*Message, error) {
	// TODO: Implement producer registration logic
	port := reg_message.port
	fmt.Printf("p = %d, t = %d\n", port, reg_message.topicID)

	var topicIdx = -1
	for i, topic := range b.topics {
		if topic.topicID == reg_message.topicID {
			topicIdx = i
			break
		}
	}

	if topicIdx == -1 {
		tp := Topic{}
		tp.init(reg_message.topicID)
		b.topics = append(b.topics, tp)
		topicIdx = len(b.topics) - 1
	}

	go func() {
		conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			fmt.Printf("Error connecting to producer at port %d: %v\n", port, err)
			return
		}
		stream_rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

		for {
			message, err := readMessageFromStream(stream_rw)
			if message == nil || err != nil {
				panic(err)
			}

			if message.PCM != nil {
				resp, err := b.processProducerPCM(message.PCM, uint16(topicIdx))
				if err != nil {
					panic(err)
				}

				err = writeMessageToStream(stream_rw, resp)
				if err != nil {
					panic(err)
				}
				continue
			}

			// Process message
			resp, err := b.processBrokerMessage(message)
			if err != nil {
				panic(err)
			}

			// Write message
			err = writeMessageToStream(stream_rw, resp)
			if err != nil {
				panic(err)
			}
		}
	}()

	var resp_byte byte = 1
	return &Message{R_P_REG: &resp_byte}, nil
}
