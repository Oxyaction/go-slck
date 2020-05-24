package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
)

var DELIMITER = []byte(`\r\n`)

type client struct {
	conn       net.Conn
	outbound   chan<- command
	register   chan<- *client
	deregister chan<- *client
	username   string
}

func (c *client) read() error {
	for {
		msg, err := bufio.NewReader(c.conn).ReadBytes('\n')
		if err == io.EOF {
			c.deregister <- c
			return nil
		}
		if err != nil {
			return err
		}
		c.handle(msg)
	}
}

func (c *client) handle(msg []byte) {
	cmd := bytes.ToUpper(bytes.TrimSpace(bytes.Split(msg, []byte(" "))[0]))
	args := bytes.TrimSpace(bytes.TrimPrefix(msg, cmd))

	switch string(cmd) {
	case "REG":
		if err := c.reg(args); err != nil {
			c.err(err)
		}
	case "MSG":
		if err := c.msg(args); err != nil {
			c.err(err)
		}
	case "JOIN":
		if err := c.join(args); err != nil {
			c.err(err)
		}
	default:
		c.err(fmt.Errorf("Unknown command '%s'\n", string(cmd)))
	}
}

func (c *client) reg(args []byte) error {
	u := bytes.TrimSpace(args)
	if len(u) == 0 {
		return errors.New("Name can not be blank")
	}
	if u[0] != '@' {
		return errors.New("Name should start with '@'")
	}

	c.username = string(u)
	c.register <- c
	return nil
}

func (c *client) msg(args []byte) error {
	args = bytes.TrimSpace(args)
	if args[0] != '@' && args[0] != '#' {
		return fmt.Errorf("Recipient should start with '@' (user) or '#' (channel), got '%s'", string(args))
	}

	recipient := bytes.Split(args, []byte(" "))[0]

	if len(recipient) == 0 {
		return fmt.Errorf("Incorrect recipient '%s'", recipient)
	}

	args = bytes.TrimSpace(bytes.TrimPrefix(args, recipient))

	l := bytes.Split(args, DELIMITER)[0]

	length, err := strconv.Atoi(string(l))
	if err != nil {
		return fmt.Errorf("body length must be present")
	}
	if length == 0 {
		return fmt.Errorf("body length must be at least 1")
	}

	padding := len(l) + len(DELIMITER)

	cmd := command{
		id:        MSG,
		recipient: string(recipient),
		sender:    c.username,
		body:      args[padding : padding+length],
	}

	c.outbound <- cmd

	return nil
}

func (c *client) join(args []byte) error {
	args = bytes.TrimSpace(args)
	if args[0] != '#' {
		return errors.New("Channel should begin with '#'")
	}

	cmd := command{
		id:        JOIN,
		recipient: string(args),
		sender:    c.username,
	}

	c.outbound <- cmd

	return nil
}

func (c *client) err(e error) {
	c.conn.Write([]byte("ERR " + e.Error() + "\n"))
}
