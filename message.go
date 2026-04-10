package main

import (
	"bufio"
	"fmt"
)

const (
	ECHO  = 1
	P_REG = 2
	// Response
	R_ECHO  = 101
	R_P_REG = 102
)

type Message struct {
	ECHO  *string
	P_REG *string
	// Response
	R_ECHO  *string
	R_P_REG *byte
}

func parseMessage(message []byte) *Message {
	switch message[0] {
	case ECHO:
		var st = string(message[1:])
		return &Message{ECHO: &st}
	case P_REG:
		var st = string(message[1:])
		return &Message{P_REG: &st}
	case R_ECHO:
		var st = string(message[1:])
		return &Message{R_ECHO: &st}
	case R_P_REG:
		var b = message[1]
		return &Message{R_P_REG: &b}
	default:
		return nil
	}
}

// Message format:
// stream[0]: size
// stream[1]: type
// stream[2...]: content
func readFromStream(stream_rw *bufio.ReadWriter) ([]byte, error) {
	var err error

	header, err := stream_rw.ReadByte() // Block
	if err != nil {
		return nil, err
	}

	data, err := stream_rw.Peek(int(header)) // Block. Data is a slice of bytes in buffer which can be modified.
	if err != nil {
		return nil, err
	}
	stream_rw.Discard(int(header))
	return data, nil
}

func readMessageFromStream(stream_rw *bufio.ReadWriter) (*Message, error) {
	data, err := readFromStream(stream_rw)
	if err != nil {
		return nil, err
	}
	return parseMessage(data), nil
}

func writeToStreamWithType(stream_rw *bufio.ReadWriter, msgType byte, data string) error {
	var err error

	// Write length
	err = stream_rw.WriteByte(byte(len(data) + 1)) // +1 for message type
	if err != nil {
		return err
	}
	// Write type
	err = stream_rw.WriteByte(msgType)
	if err != nil {
		return err
	}

	// Write content
	_, err = stream_rw.WriteString(data)
	if err != nil {
		return err
	}

	// Flush
	err = stream_rw.Flush()
	if err != nil {
		return err
	}

	return nil
}

func writeMessageToStream(stream_rw *bufio.ReadWriter, message *Message) error {
	if message.ECHO != nil {
		return writeToStreamWithType(stream_rw, ECHO, *message.ECHO)
	} else if message.P_REG != nil {
		return writeToStreamWithType(stream_rw, P_REG, *message.P_REG)
	} else if message.R_ECHO != nil {
		return writeToStreamWithType(stream_rw, R_ECHO, *message.R_ECHO)
	} else if message.R_P_REG != nil {
		data := fmt.Sprintf("%d", *message.R_P_REG)
		return writeToStreamWithType(stream_rw, R_P_REG, data)
	}
	return nil
}
