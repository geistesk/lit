package main

import (
	"fmt"
	"strings"
)

type Message struct {
	Channel  string
	User     string
	Location string
	Text     string
}

func ParseMessage(msg string) (m Message, err error) {
	parts := strings.Split(msg, "|")
	if len(parts) < 4 {
		err = fmt.Errorf("expected >= 4 parts; got %d parts", len(parts))
		return
	}

	m = Message{
		Channel:  parts[0],
		User:     parts[1],
		Location: parts[2],
		Text:     strings.Join(parts[3:], "|"),
	}
	return
}

func (m Message) String() string {
	loc := m.Location
	if loc == "" {
		loc = "0.0,0.0"
	}

	return fmt.Sprintf("%s|%s|%s|%s", m.Channel, m.User, loc, m.Text)
}
