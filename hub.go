package main

import (
	"errors"
	"fmt"
	"strings"
)

type hub struct {
	channels        map[string]*channel
	clients         map[string]*client
	commands        chan command
	registrations   chan *client
	deregistrations chan *client
}

func NewHub() *hub {
	return &hub{
		channels:        make(map[string]*channel),
		clients:         make(map[string]*client),
		commands:        make(chan command),
		registrations:   make(chan *client),
		deregistrations: make(chan *client),
	}
}

func (h *hub) run() {
	for {
		select {
		case client := <-h.registrations:
			h.register(client)
		case client := <-h.deregistrations:
			h.deregister(client)
		case cmd := <-h.commands:
			switch cmd.id {
			case JOIN:
				h.joinChannel(cmd.sender, cmd.recipient)
			case MSG:
				h.sendMessage(cmd.sender, cmd.recipient, cmd.body)
			}
		}
	}
}

func (h *hub) register(c *client) {
	if _, exists := h.clients[c.username]; exists {
		c.username = ""
		c.err(errors.New("User already exists"))
	} else {
		h.clients[c.username] = c
		c.conn.Write([]byte("OK\n"))
	}
}

func (h *hub) deregister(c *client) {
	if _, exists := h.clients[c.username]; exists {
		delete(h.clients, c.username)
		for _, ch := range h.channels {
			delete(ch.clients, c)
		}
	}

	c.conn.Write([]byte("OK\n"))
}

func (h *hub) joinChannel(u, c string) {
	if cl, ok := h.clients[u]; ok {
		ch, ok := h.channels[c]
		if !ok {
			ch = &channel{
				name:    c,
				clients: make(map[*client]bool),
			}
			h.channels[ch.name] = ch
		}
		ch.clients[cl] = true

		cl.conn.Write([]byte("OK\n"))
	}

}

func (h *hub) sendMessage(s, r string, msg []byte) {
	if cl, ok := h.clients[s]; ok {
		msg = append(msg, "\n"...)
		if strings.HasPrefix(r, "@") {
			if client, ok := h.clients[r]; ok {
				client.conn.Write(append([]byte(cl.username+": "), msg...))
			}
		} else {
			fmt.Println(h.channels)
			if channel, ok := h.channels[r]; ok {
				channel.broadcast(cl.username, msg)
			}
		}
	}
}
