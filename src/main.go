package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/porty/emitter"
)

type Events struct {
	TransmitTime time.Time       `json:"time"`
	Events       []emitter.Event `json:"events"`
}

var events *Events

func addEvent(e emitter.Event) {
	events.Events = append(events.Events, e)
}

func actuallySendEvents(e *Events) {
	e.TransmitTime = time.Now()

	b, err := json.Marshal(e)
	if err != nil {
		fmt.Println("Failed to JSON Events: " + err.Error())
		return
	}

	for i := 0; i < 3; i++ {
		resp, err := http.Post("http://localhost:3000/events", "application/json", bytes.NewReader(b))

		if err == nil {
			if resp.StatusCode >= 300 {
				fmt.Println("Server returned bad response: " + resp.Status)
			}
			fmt.Println("Sent good :)")
			return
		}

		fmt.Println("Failed to post events: " + err.Error())
	}
}

func sendEvents() {
	if len(events.Events) > 0 {
		fmt.Println("I have events to send!")
		eventsToSend := events
		events = &Events{}

		go actuallySendEvents(eventsToSend)
	} else {
		fmt.Println("Nothing to send :(")
	}
}

func loopies(c chan emitter.Event) {
	for {
		t := time.NewTicker(5 * time.Second)

		select {
		case e := <-c:
			addEvent(e)
		case <-t.C:
			sendEvents()
		}
	}
}

func main() {
	events = &Events{}
	a, err := emitter.NewAbsorber()
	if err != nil {
		panic(err)
	}
	c := make(chan emitter.Event)
	go loopies(c)

	for {
		fmt.Println("Waiting...")
		e, err := a.WaitForEvent()
		if err != nil {
			fmt.Println("Failed event: " + err.Error())
		}
		fmt.Println("OMG i has event")

		c <- *e

	}
}
