package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	turboumbrella "github.com/lemon-mint/turbo-umbrella"
)

func greet(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "Hello From New")
}

func main() {
	log.Println("My PID:", os.Getpid())

	flag.Parse()
	t, err := turboumbrella.New("myhttp", "tcp", ":8080")
	if err != nil {
		panic(err)
	}
	s := &http.Server{
		IdleTimeout: time.Second * 10,
	}
	t.OnUpgrade = func() {
		log.Println("Shutting down Server...")
		err := s.Close()
		if err != nil {
			log.Println(err)
		}
		log.Println("Server Closed")
	}
	s.Handler = http.HandlerFunc(greet)

	go func() {
		time.Sleep(time.Second)
		fmt.Println("Upgrading...")
		err = t.Upgrade(time.Second)
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Second * 2)

		go func() {
			fmt.Println(t.WaitForUpgrade())
		}()
	}()

	log.Println("Listening...")
	err = s.Serve(t.Listener())
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
