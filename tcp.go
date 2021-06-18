package tcpchan

import (
	"bufio"
	"fmt"
	"go.uber.org/zap"
	"net"
	"strings"
)


func (c *Connection) connect() (err error) {
	log.Debug("Wait lock for (re)connection")
	c.isFinishedLock.Lock()
	defer c.isFinishedLock.Unlock()
	log.Debug("Create a new connection")
	c.conn, err = net.Dial("tcp4", c.address)
	if err != nil {
		return fmt.Errorf("error while establing connection: %w", err)
	}
	log.Debug("Read headlines")
	c.reader = bufio.NewReader(c.conn)
	c.readline()
	log.Debug("TCP connected")
	return nil
}

func (c *Connection) closeTCP() {
	log.Debug("Close old connection (if exists)")
	if c.reader != nil {
		c.reader = nil
	}
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			log.Error("Error while closing connection", zap.String("error", err.Error()))
		}
		c.conn = nil
	}
}

var log *zap.Logger

func init() {
	{
		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		log, _ = cfg.Build()
	}
}


func (c *Connection) readline() {
	defer func() {
		if r := recover(); r != nil {
			if c.isFinished {
				log.Warn("Ignore panic because connection is closing.")
			} else {
				panic(r)
			}
		}
	}()

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
			log.Error(err.Error())
			panic(fmt.Errorf("read line error: %w", err))
		}
	}

	c.Data <- string(line) // where to close?
}

