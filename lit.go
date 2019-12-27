package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

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

	// logfile
	logfile = "lora_rec.csv"
)

var (
	irccon *irc.Connection
	modem  *rf95.Modem
	logger *os.File
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
			fmt.Printf("error: %v\n", err)
		}
	})

	if err := irccon.Connect(server); err != nil {
		panic(err)
	}

	go irccon.Loop()
}

func setupLoRaLogfile() {
	if f, fErr := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); fErr != nil {
		panic(fErr)
	} else {
		logger = f
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
		csvLog := fmt.Sprintf("%d,%x,%d,%d\n", time.Now().UnixNano(), rxMessage.Payload, rxMessage.Rssi, rxMessage.Snr)
		if _, err := logger.WriteString(csvLog); err != nil {
			fmt.Printf("error: %v\n", err)
		}
		if err := logger.Sync(); err != nil {
			fmt.Printf("error: %v\n", err)
		}

		if msg, err := ParseMessage(string(rxMessage.Payload)); err == nil {
			if msg.Channel != "36c3" {
				irccon.Privmsgf(channel, "#%s <%s> %s", msg.Channel, msg.User, msg.Text)
			} else {
				irccon.Privmsgf(channel, "<%s> %s", msg.User, msg.Text)
			}
		}
	})
}

func waitSigint() {
	signalSyn := make(chan os.Signal)
	signalAck := make(chan struct{})

	signal.Notify(signalSyn, os.Interrupt)

	go func() {
		<-signalSyn
		close(signalAck)
	}()

	<-signalAck
}

func main() {
	setupIrc()
	setupLoRaLogfile()
	setupLoRa()

	waitSigint()

	_ = modem.Close()
	_ = logger.Close()
	irccon.Quit()
}
