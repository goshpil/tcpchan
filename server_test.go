package tcpchan

import (
	"bufio"
	"io"
	"net"
	"testing"
)

const ServerAddress = "localhost:9999"

func encrypt(s []byte) {
	for i, v := range s {
		s[i] = v + 1
	}
}

func runServer(t *testing.T) {
	t.Helper()
	ready := make(chan struct{})

	go func() {
		l, err := net.Listen("tcp", ServerAddress)
		if err != nil {
			t.Fatalf("Error listening: %v", err)
		}
		defer l.Close()
		t.Log("[Server] TCP Mock server started")
		ready <- struct{}{}

		for {
			conn, err := l.Accept()
			if err != nil {
				t.Fatalf("Error accepting: %v", err)
			}
			go func(conn net.Conn) {
				_, err := conn.Write([]byte("HELLO\n"))
				if err != nil {
					t.Fatalf("Error writing hello: %v", err)
				}
				reader := bufio.NewReader(conn)
				for {
					t.Log("[Server] Wait for client message")
					line, isPrefix, err := reader.ReadLine()
					if err == io.EOF {
						t.Log("[Server] EOF")
						return
					} else if err != nil {
						t.Fatalf("Error reading: %v", err)
					}
					if isPrefix {
						t.Fatalf("Is prefix")
					}
					response := make([]byte, len(line))
					copy(response, line)
					encrypt(response)
					t.Logf("[Server] Received: %s Sent: %s", string(line), string(response))
					_, err = conn.Write(append(response, '\n'))
					if err != nil {
						t.Fatalf("Error writing response: %v", err)
					}
				}
			}(conn)
		}
	}()

	<-ready
}
