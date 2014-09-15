package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	postURL:      "http://lenny.compassuav.com/events",
	postInterval: 5 * time.Second,
	tryCount:     3,
	eventCap:     50,
}

type Events struct {
	TransmitTime time.Time       `json:"time"`
	Events       []emitter.Event `json:"events"`
}

var events *Events

var droppingEvents = false

func addEvent(e emitter.Event) {
	if len(events.Events) < s.eventCap {
		events.Events = append(events.Events, e)
	} else if !droppingEvents {
		droppingEvents = true
		fmt.Println("Too many events! They are being dropped!")
	}

}

func actuallySendEvents(e *Events) {
	e.TransmitTime = time.Now()

	b, err := json.Marshal(e)
	if err != nil {
		fmt.Println("Failed to marshal events to JSON: " + err.Error())
		return
	}

	for i := 0; i < s.tryCount; i++ {
		resp, err := http.Post(s.postURL, "application/json", bytes.NewReader(b))

		if err == nil {
			if resp.StatusCode >= 300 {
				fmt.Println("Server returned bad response: " + resp.Status)
			}
			fmt.Printf("Transmitted %d events (%d bytes)\n", len(e.Events), len(b))
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
		droppingEvents = false

		go actuallySendEvents(eventsToSend)
	}
}

func loopies(c chan emitter.Event) {
	t := time.NewTicker(s.postInterval)

	for {
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

	fmt.Printf("ECAT ready.\nTransmitting events every %d seconds to %s\n",
		int(s.postInterval.Seconds()),
		s.postURL,
	)

	for {
		e, err := a.WaitForEvent()
		if err != nil {
			fmt.Println("Failed event: " + err.Error())
		}

		c <- *e

	}
}
