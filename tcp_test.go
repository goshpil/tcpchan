package tcpchan

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStableTCP(t *testing.T) {
	runServer(t)
	c := NewConnection(ServerAddress)
	c.Start()
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
	require.Equal(t, c.isFinished, true)
	require.Error(t, c.Send([]byte("AFTERCLOSE\n")))
}
