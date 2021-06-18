package tcpchan

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"
)

type Hook func()

type Connection struct {
	address string

	conn   net.Conn
	reader *bufio.Reader
	Data   chan string

	// graceful start
	ready           chan struct{}

	// graceful shutdown
	isFinished     bool
	isFinishedLock sync.RWMutex
	waitSends      sync.WaitGroup
	waitReader     sync.WaitGroup
	stopReader     context.CancelFunc
}

func NewConnection(address string) *Connection {
	c := &Connection{
		address: address,
		Data:    make(chan string, 1000),
		ready:   make(chan struct{}),
	}
	return c
}

func (c *Connection) Start() {
	log.Debug("TCP connection keeper started")
	defer log.Debug("TCP connection keeper stopped")

	var ctx context.Context
	ctx, c.stopReader = context.WithCancel(context.Background())
	c.waitReader.Add(1)

	log.Debug("Run working loop")
	go c.runWorkingLoop(ctx)

	log.Debug("Wait for TCP connection established")
	<-c.ready
	log.Info("TCP connection started")
}

func (c *Connection) Close() {
	log.Debug("Stop get new Send function calls")
	c.isFinishedLock.Lock()
	c.isFinished = true
	c.isFinishedLock.Unlock()

	log.Debug("Wait while existent Send function calls finished")
	c.waitSends.Wait()

	log.Debug("Close connection")
	c.stopReader()
	close(c.Data)
	c.closeTCP()

	log.Debug("Graceful shutdown")
}

func (c *Connection) Send(message []byte) error {
	c.isFinishedLock.RLock()
	defer c.isFinishedLock.RUnlock()
	if c.isFinished {
		return fmt.Errorf("connection already closed")
	}
	c.waitSends.Add(1)
	defer c.waitSends.Done()

	_, err := c.conn.Write(message)
	if err != nil {
		return fmt.Errorf("cannot send tcp message to tcp server: %w", err)
	}
	return nil
}
