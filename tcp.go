package tcpchan

import (
	"bufio"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"strings"
	"sync"
	"time"
)

type Hook func()

type Connection struct {
	address string

	conn   net.Conn
	reader *bufio.Reader
	Data   chan string

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
	}
	return c
}

func (c *Connection) connect() (err error) {
	c.closeTCP()
	c.conn, err = net.Dial("tcp4", c.address)
	if err != nil {
		return fmt.Errorf("error while establing connection: %w", err)
	}
	c.reader = bufio.NewReader(c.conn)
	c.readline()
	return nil
}

func (c *Connection) closeTCP() {
	logrus.Debug("Close old connection (if exists)")
	if c.reader != nil {
		c.reader = nil
	}
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			logrus.WithError(err).Error("Error while closing connection")
		}
		c.conn = nil
	}
}

func (c *Connection) Start() {
	logrus.Debug("TCP connection keeper started")
	defer logrus.Debug("TCP connection keeper stopped")

	var ctx context.Context
	ctx, c.stopReader = context.WithCancel(context.Background())
	c.waitReader.Add(1)
	defer c.waitReader.Done()
	defer c.closeTCP()

	err := c.connect()
	logrus.Info("TCP connection started")

	for {
		select {
		case <-ctx.Done():
			logrus.Debug("TCP connection context is done 1")
			return
		default:
			if err != nil {
				logrus.WithError(err).Error("TCP connection error occurred. Need reconnect.")
				select {
				case <-ctx.Done():
					logrus.Debug("TCP connection context is done 2")
					return
				case <-time.After(time.Second):
					err = c.connect()
					logrus.Warn("TCP connection restarted")
					continue
				}
			}
			c.readline()
		}
	}
}

func (c *Connection) Close() {
	logrus.Debug("Stop get new Send function calls")
	c.isFinishedLock.Lock()
	c.isFinished = true
	c.isFinishedLock.Unlock()

	logrus.Debug("Wait while existent Send function calls finished")
	c.waitSends.Wait()

	logrus.Debug("Close connection")
	c.stopReader()
	c.closeTCP()
	close(c.Data)

	logrus.Debug("Graceful shutdown")
}

func (c *Connection) readline() {
	line, isPrefix, err := c.reader.ReadLine()

	if isPrefix {
		panic("reader buffer is too small")
	}
	if err != nil {
		if strings.HasSuffix(err.Error(), ": use of closed network connection") {
			c.isFinishedLock.RLock()
			defer c.isFinishedLock.RUnlock()
			if !c.isFinished {
				panic(fmt.Errorf("unexpected closed network connection"))
			}
		} else {
			logrus.Error(err.Error())
			panic(fmt.Errorf("read line error: %w", err))
		}
	}

	c.Data <- string(line) // where to close?
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
