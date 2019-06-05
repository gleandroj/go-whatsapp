package main

import (
	"fmt"
	"github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/gleandroj/go-whatsapp"
	"os"
	"time"
)

func main() {
	wac, err := whatsapp.NewConn(5 * time.Second, "gleandroj/go-whatsapp", "go-whatsapp")
	if err != nil {
		panic(err)
	}

	qr := make(chan string)
	go func() {
		terminal := qrcodeTerminal.New()
		terminal.Get(<-qr).Print()
	}()

	session, err := wac.Login(qr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during login: %v\n", err)
	}
	fmt.Printf("login successful, session: %v\n", session)
}
