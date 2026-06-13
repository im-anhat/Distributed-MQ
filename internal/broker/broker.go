package broker

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/im-anhat/Distributed-MQ/internal/config"
	"github.com/im-anhat/Distributed-MQ/internal/topic"
	"github.com/im-anhat/Distributed-MQ/internal/wire"
)

type Broker struct {
	topics []topic.Topic
}

func (b *Broker) Init() {
	b.topics = make([]topic.Topic, 0)
}

// bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
func (b *Broker) StartBrokerServer() error {
	ln, _ := net.Listen("tcp", fmt.Sprintf(":%d", config.BrokerPort))
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

// Head will not advance until we pop. Trims messages all groups have committed past.
func (b *Broker) StopAndPop(topicIdx uint) {
	for {
		time.Sleep(5 * time.Second)
		minOffset := -1
		for i, cgroup := range b.topics[topicIdx].Cgroups {
			if int(cgroup.Offset) < minOffset || minOffset == -1 {
				minOffset = int(cgroup.Offset)
			}
			b.topics[topicIdx].Cgroups[i].Lock.Lock()
		}
		if minOffset == -1 {
			for i := range b.topics[topicIdx].Cgroups {
				b.topics[topicIdx].Cgroups[i].Lock.Unlock()
			}
			continue
		}
		totalPop := minOffset
		fmt.Printf("Stop and pop running, minOffset = %d\n", minOffset)
		for {
			if minOffset == 0 {
				break
			}
			b.topics[topicIdx].MQ.Pop()
			minOffset--
		}
		for i := range b.topics[topicIdx].Cgroups {
			b.topics[topicIdx].Cgroups[i].Offset -= uint(totalPop)
			b.topics[topicIdx].Cgroups[i].Lock.Unlock()
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
	if message.C_REG != nil {
		resp, err = b.processConsumerRegisterMessage(message.C_REG)
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
		go b.StopAndPop(uint(topicIdx))
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

func (b *Broker) processConsumerRegisterMessage(reg_message *wire.ConsumerRegisterMessage) (*wire.Message, error) {
	fmt.Printf("Broker received consumer registration: port=%d, topicID=%d, groupID=%d\n", reg_message.Port, reg_message.TopicID, reg_message.GroupID)

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

	var groupIdx = -1
	for i, group := range b.topics[topicIdx].Cgroups {
		if group.GroupID == reg_message.GroupID {
			groupIdx = i
			break
		}
	}
	if groupIdx == -1 {
		group := topic.CGroup{
			GroupID: reg_message.GroupID,
			Offset:  0,
		}
		b.topics[topicIdx].Cgroups = append(b.topics[topicIdx].Cgroups, group)
		groupIdx = len(b.topics[topicIdx].Cgroups) - 1
		go b.startConsumerGroupConsumption(uint(topicIdx), uint(groupIdx))
	}

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", reg_message.Port))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Connected to Consumer at port %d\n", reg_message.Port)

	var consumer = topic.ConsumerConn{
		Status: true,
		Conn:   conn,
	}
	b.topics[topicIdx].Cgroups[groupIdx].Consumers = append(b.topics[topicIdx].Cgroups[groupIdx].Consumers, consumer)

	var resp_byte byte = 1
	return &wire.Message{R_C_REG: &resp_byte}, nil
}

func (b *Broker) startConsumerGroupConsumption(topicIdx uint, cgroupIdx uint) {
	for {
		b.topics[topicIdx].Cgroups[cgroupIdx].Lock.Lock()
		offset := b.topics[topicIdx].Cgroups[cgroupIdx].Offset
		pcm := b.topics[topicIdx].MQ.Peek(offset)
		b.topics[topicIdx].Cgroups[cgroupIdx].Lock.Unlock()
		if len(pcm) == 0 {
			fmt.Printf("No PCM to consume at offset %d\n", offset)
			time.Sleep(1 * time.Second)
			continue
		}

		for _, consumer := range b.topics[topicIdx].Cgroups[cgroupIdx].Consumers {
			if consumer.Status {
				conn := consumer.Conn
				stream_rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

				// Write
				consumer.Status = false
				err := wire.WriteMessageToStream(stream_rw, &wire.Message{PCM: pcm})
				if err != nil {
					panic(err)
				}

				// Read ack
				parsed_message, err := wire.ReadMessageFromStream(stream_rw)
				if parsed_message == nil || err != nil {
					panic(err)
				}
				if parsed_message.R_PCM != nil {
					consumer.Status = true
					b.topics[topicIdx].Cgroups[cgroupIdx].Lock.Lock()
					b.topics[topicIdx].Cgroups[cgroupIdx].Offset++
					b.topics[topicIdx].Cgroups[cgroupIdx].Lock.Unlock()
				}
			}
		}
	}
}
