package main

type channel struct {
	name    string
	clients map[*client]bool
}

func (c *channel) broadcast(s string, message []byte) {
	msg := append([]byte(s), ": "...)
	msg = append(msg, message...)
	for client := range c.clients {
		client.conn.Write(msg)
	}
}
