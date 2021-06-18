# tcpchan

A simple TCP listener proxying TCP messages to a golang channel.

```golang
c := NewConnection("localhost:9999")

go func() {
  time.Sleep(time.Second)
  c.Close()
}()

for msg := range c.Data {
  fmt.Println(msg)
}
``` 
