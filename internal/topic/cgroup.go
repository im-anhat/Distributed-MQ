package topic

import (
	"net"
	"sync"
)

type CGroup struct {
	GroupID   uint16
	Offset    uint
	Consumers []ConsumerConn
	Lock      sync.Mutex
}

type ConsumerConn struct {
	Status bool
	Conn   net.Conn
}
