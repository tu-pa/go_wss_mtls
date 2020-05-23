package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"crypto/x509"
	"crypto/tls"

	"os"
	"os/signal"
	"flag"
	"net"
	"net/url"
	"time"
	"math/rand"

	"github.com/gorilla/websocket"
)

var serverAddr = flag.String("addr", "127.0.0.1:8443", "http service address")
var udpPort = ":8444"
func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Create a CA certificate pool and add cert.pem to it
	caCert, err := ioutil.ReadFile("certs/cert.pem")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Read the key pair to create certificate
	cert, err := tls.LoadX509KeyPair("certs/cert.pem", "certs/key.pem")
	if err != nil {
		log.Fatal(err)
	}

	u := url.URL{Scheme: "wss", Host: *serverAddr, Path: "/echo"}
	log.Printf("connecting to %s", u.String())

	wssDialer := websocket.DefaultDialer
	wssDialer.TLSClientConfig = &tls.Config{
		RootCAs: caCertPool,
		Certificates: []tls.Certificate{cert},
	}

	c, _, err := wssDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	// UDP Listen
	s, err := net.ResolveUDPAddr("udp4", udpPort)
	if err != nil {
			fmt.Println(err)
			return
	}

	udpCon, err := net.ListenUDP("udp4", s)
	if err != nil {
			fmt.Println(err)
			return
	}
	defer udpCon.Close()
	buffer := make([]byte, 1024)
	rand.Seed(time.Now().Unix())

	go func () {
		defer close(done)
		for {
			udpCon.ReadFromUDP(buffer)
			//fmt.Print("-> ", string(buffer[0:n-1]))
			//case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, buffer)
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	} ()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}