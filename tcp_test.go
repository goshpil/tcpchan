package tcpchan

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestStableTCP(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	runServer(t)
	c := NewConnection(ServerAddress)
	go c.Start()
	time.Sleep(time.Second)

	hello := <-c.Data
	require.Equal(t, "HELLO", hello)

	for i := 1; i <= 100; i++ {
		message := getRandomMessage(i)
		t.Logf("[Client] Send message: %s", message)
		err := c.Send(append(message, '\n'))
		require.NoErrorf(t, err, "Error sending message")

		expected := make([]byte, len(message))
		copy(expected, message)
		encrypt(expected)

		t.Log("[Client] Wait response")
		actual := <-c.Data
		t.Logf("[Client] Response: %s", actual)
		require.Equal(t, string(expected), actual, "Wrong response")
	}

	c.Close()
}
