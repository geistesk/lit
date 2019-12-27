package main

import (
	"github.com/dtn7/rf95modem-go/rf95"
	irc "github.com/thoj/go-ircevent"
)

const (
	// IRC settings
	nick    = "litbot"
	user    = "lit"
	server  = "irc.hackint.org:6697"
	channel = "#lit"

	// rf95modem settings
	device    = "/dev/ttyUSB0"
	freq      = 868.1
	modemMode = rf95.MediumRange
)

var (
	irccon *irc.Connection
	modem  *rf95.Modem
)

func setupIrc() {
	irccon = irc.IRC(nick, user)
	irccon.UseTLS = true

	irccon.AddCallback("001", func(e *irc.Event) { irccon.Join(channel) })
	irccon.AddCallback("PRIVMSG", func(e *irc.Event) {
		msg := Message{
			Channel:  "36c3",
			User:     e.Nick,
			Location: "",
			Text:     e.Message(),
		}

		if _, err := modem.Write([]byte(msg.String())); err != nil {
			panic(err)
		}
	})

	if err := irccon.Connect(server); err != nil {
		panic(err)
	}
}

func setupLoRa() {
	if m, mErr := rf95.OpenSerial(device); mErr != nil {
		panic(mErr)
	} else {
		modem = m
	}

	if freqErr := modem.Frequency(freq); freqErr != nil {
		panic(freqErr)
	}

	if modeErr := modem.Mode(modemMode); modeErr != nil {
		panic(modeErr)
	}

	modem.RegisterRxHandler(func(rxMessage rf95.RxMessage) {
		if msg, err := ParseMessage(string(rxMessage.Payload)); err == nil {
			irccon.Privmsgf(channel, "#%s <%s> %s", msg.Channel, msg.User, msg.Text)
		}
	})
}

func main() {
	setupIrc()
	setupLoRa()

	irccon.Loop()
}
