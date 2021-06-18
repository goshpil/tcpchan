package tcpchan

import (
	"context"
	"go.uber.org/zap"
	"time"
	"fmt"
)

func (c *Connection)runWorkingLoop(ctx context.Context) {
	defer c.waitReader.Done()

	log.Debug("Start first TCP connection")
	err := c.connect()
	for err != nil {
		log.Error("TCP first connection error", zap.Error(err))
		err = c.connect()
	}
	c.ready <- struct{}{}
	log.Debug("TCP connection keeper marked as ready")
	fmt.Println("TCP connection keeper marked as ready")

	for {
		select {
		case <-ctx.Done():
			log.Debug("TCP connection context is done (while iteration)")
			return
		default:
			if err != nil {
				log.Error("TCP connection error occurred. Need reconnect.", zap.Error(err))
				select {
				case <-ctx.Done():
					log.Debug("TCP connection context is done (while sleeping)")
					return
				case <-time.After(time.Second):
					err = c.connect()
					log.Warn("TCP connection restarted")
					continue
				}
			}
			log.Debug("Read next line")
			c.readline()
		}
	}
}
