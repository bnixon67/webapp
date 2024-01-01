package main

import (
	"log"
	"net/http"
	"time"

	"github.com/bnixon67/webapp/websse"
)

func main() {
	s := websse.NewServer()
	s.RegisterEvents("", "event1", "event2")
	s.Run()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "assets/htmx.html")
	})

	http.HandleFunc("/event", s.EventStreamHandler)

	go func() {
		err := http.ListenAndServeTLS(":8443", "cert/cert.pem", "cert/key.pem", nil)
		if err != nil {
			log.Fatal("ListenAndServeTLS: ", err)
		}
	}()

	for {
		s.Publish(websse.Message{Event: "event1", Data: "data"})
		time.Sleep(1 * time.Second)
	}

	select {}
}
