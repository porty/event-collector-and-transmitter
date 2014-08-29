package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/porty/emitter"
)

type settings struct {
	postURL      string
	postInterval time.Duration
	tryCount     int
	eventCap     int
}

var s = settings{
	postURL:      "http://localhost:3000/events",
	postInterval: 5 * time.Second,
	tryCount:     3,
	eventCap:     50,
}

type Events struct {
	TransmitTime time.Time       `json:"time"`
	Events       []emitter.Event `json:"events"`
}

var events *Events

func addEvent(e emitter.Event) {
	events.Events = append(events.Events, e)
}

func actuallySendEvents(e *Events) {

	if len(e.Events) > s.eventCap {
		log.Printf("Found %d events to send, trimming to %d", len(e.Events), s.eventCap)
		e.Events = e.Events[:s.eventCap]
	}

	e.TransmitTime = time.Now()

	b, err := json.Marshal(e)
	if err != nil {
		fmt.Println("Failed to JSON Events: " + err.Error())
		return
	}

	for i := 0; i < s.tryCount; i++ {
		resp, err := http.Post(s.postURL, "application/json", bytes.NewReader(b))

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
		// allocate new Events struct so that the goroutine has its own copy
		eventsToSend := events
		events = &Events{}

		go actuallySendEvents(eventsToSend)
	}
}

func loopies(c chan emitter.Event) {
	for {
		t := time.NewTicker(s.postInterval)

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
