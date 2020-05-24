package main

type ID int

const (
	REG ID = iota
	JOIN
	LEAVE
	MSG
)

type command struct {
	id        ID
	recipient string
	sender    string
	body      []byte
}
