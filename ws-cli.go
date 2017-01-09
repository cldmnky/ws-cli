package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/chzyer/readline"
	ws "github.com/gorilla/websocket"
)

func recv(conn *ws.Conn, rl *readline.Instance, wg *sync.WaitGroup, interrupt func()) {
	defer func() {
		interrupt()
		wg.Done()
	}()

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			break
		}
		fmt.Fprintf(rl.Stdout(), "< %s\n", string(p))
	}
}

func send(conn *ws.Conn, rl *readline.Instance, wg *sync.WaitGroup) {
	defer func() {
		conn.Close()
		wg.Done()
	}()

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			break
		} else if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("ReadLine: %v", err)
		}

		// Remove the new line from the string.
		conn.WriteMessage(ws.TextMessage, []byte(line))
	}
}

func dial(url string, origin string, extraHeader string, subprotocol string) (*ws.Conn, error) {
	var subprotocols []string
	var header http.Header

	if subprotocol != "" {
		subprotocols = []string{subprotocol}
	}
	if origin != "" {
		header = http.Header{"Origin": {origin}}
	}

	if extraHeader != "" {
		h := strings.Split(extraHeader, ":")
		log.Printf("Adding header: %s:%s", h[0], strings.Trim(h[1], " "))
		header = http.Header{h[0]: {h[1]}}
	}

	dialer := ws.Dialer{
		Subprotocols:    subprotocols,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, _, err := dialer.Dial(url, header)
	return conn, err
}

func main() {
	var url = flag.String("url", "", "url")
	var origin = flag.String("origin", "", "optional origin")
	var subprotocol = flag.String("subprotocol", "", "optional subprotocol")
	var extraHeader = flag.String("header", "", "optional extra header")
	flag.Parse()
	if *url == "" {
		flag.Usage()
		return
	}

	conn, err := dial(*url, *origin, *extraHeader, *subprotocol)
	if err != nil {
		log.Fatalf("Dial: %v", err)
	}
	fmt.Println("connected (press CTRL+C to quit)")

	stdin := newInterruptibleStdin(os.Stdin)
	rl, err := readline.NewEx(&readline.Config{
		Prompt: "> ",
		Stdin:  stdin,
	})
	if err != nil {
		log.Fatalf("New: %v", err)
	}
	defer rl.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go recv(conn, rl, &wg, stdin.interrupt)
	go send(conn, rl, &wg)

	wg.Wait()

	fmt.Println("Disconnected")
}
