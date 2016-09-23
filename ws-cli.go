package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/chzyer/readline"
	ws "github.com/gorilla/websocket"
)

func recv(conn *ws.Conn, rl *readline.Instance, wg *sync.WaitGroup) {
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			wg.Done()
			break
		}
		buf := new(bytes.Buffer)
		buf.WriteString("< ")
		buf.Write(p)
		buf.WriteRune('\n')

		rl.Stdout().Write(buf.Bytes())
	}
}

func send(conn *ws.Conn, rl *readline.Instance) {
	defer conn.Close()

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("ReadLine: %v", err)
		}

		// Remove the new line from the string.
		conn.WriteMessage(ws.TextMessage, []byte(line))
	}
}

func main() {
	var url = flag.String("url", "", "url")
	var subprotocol = flag.String("subprotocol", "", "optional subprotocol")
	flag.Parse()
	if *url == "" {
		flag.Usage()
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	var subprotocols []string
	if *subprotocol != "" {
		subprotocols = []string{*subprotocol}
	}
	dialer := ws.Dialer{
		Subprotocols:    subprotocols,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, _, err := dialer.Dial(*url, nil)
	if err != nil {
		log.Fatalf("Dial: %v", err)
	}
	fmt.Println("connected (press CTRL+C to quit)")

	rl, err := readline.New("> ")
	if err != nil {
		log.Fatalf("New: %v", err)
	}
	defer rl.Close()

	go recv(conn, rl, &wg)
	send(conn, rl)

	wg.Wait()
}
