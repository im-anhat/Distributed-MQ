package broker

import (
	"bufio"
	"fmt"
	"net"

	"github.com/im-anhat/Distributed-MQ/internal/topic"
	"github.com/im-anhat/Distributed-MQ/internal/wire"
)

const BROKER_PORT = 10000

type Broker struct {
	topics []topic.Topic
}

func (b *Broker) Init() {
	b.topics = make([]topic.Topic, 0)
}

// bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
func (b *Broker) StartBrokerServer() error {
	ln, _ := net.Listen("tcp", fmt.Sprintf(":%d", BROKER_PORT))
	fmt.Println("Server started...")
	for {
		conn, _ := ln.Accept() // Block
		stream_rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

		message, err := wire.ReadMessageFromStream(stream_rw)
		if err == nil && message != nil {
			resp, err := b.processBrokerMessage(message)
			if err != nil {
				return err
			}

			// Write it back
			err = wire.WriteMessageToStream(stream_rw, resp)
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

func (b *Broker) processBrokerMessage(message *wire.Message) (*wire.Message, error) {
	var err error
	var resp *wire.Message

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

func (b *Broker) processProducerPCM(pcm []byte, topicIdx uint16) (*wire.Message, error) {
	b.topics[topicIdx].MQ.Push(pcm)
	b.topics[topicIdx].MQ.Debug()
	one := byte(1)
	return &wire.Message{R_PCM: &one}, nil
}

func (b *Broker) processEchoMessage(echo_message *string) (*wire.Message, error) {
	fmt.Printf("Received Echo message: %s!", *echo_message)
	resp_echo := fmt.Sprintf("I have received your message: %s", *echo_message)
	return &wire.Message{R_ECHO: &resp_echo}, nil
}

func (b *Broker) processProducerRegisterMessage(reg_message *wire.ProducerRegisterMessage) (*wire.Message, error) {
	// TODO: Implement producer registration logic
	port := reg_message.Port
	fmt.Printf("p = %d, t = %d\n", port, reg_message.TopicID)

	var topicIdx = -1
	for i, topic := range b.topics {
		if topic.TopicID == reg_message.TopicID {
			topicIdx = i
			break
		}
	}

	if topicIdx == -1 {
		tp := topic.Topic{}
		tp.Init(reg_message.TopicID)
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
			message, err := wire.ReadMessageFromStream(stream_rw)
			if message == nil || err != nil {
				panic(err)
			}

			if message.PCM != nil {
				resp, err := b.processProducerPCM(message.PCM, uint16(topicIdx))
				if err != nil {
					panic(err)
				}

				err = wire.WriteMessageToStream(stream_rw, resp)
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
			err = wire.WriteMessageToStream(stream_rw, resp)
			if err != nil {
				panic(err)
			}
		}
	}()

	var resp_byte byte = 1
	return &wire.Message{R_P_REG: &resp_byte}, nil
}
