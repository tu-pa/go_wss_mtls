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

var controlChannel = flag.String("caddr", "127.0.0.1:8442", "control channel")
var dataChannel = flag.String("daddr", "127.0.0.1:8443", "data channel")
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

	ccUrl := url.URL{Scheme: "wss", Host: *controlChannel, Path: "/control"}
	log.Printf("connecting control channel to %s", ccUrl.String())
	dcUrl := url.URL{Scheme: "wss", Host: *dataChannel, Path: "/control"}
	log.Printf("connecting data channel to %s", dcUrl.String())

	wssDialer := websocket.DefaultDialer
	wssDialer.TLSClientConfig = &tls.Config{
		RootCAs: caCertPool,
		Certificates: []tls.Certificate{cert},
	}

	cc, _, err := wssDialer.Dial(ccUrl.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer cc.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			// TODO - Server Control
			_, message, err := cc.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	dc, _, err := wssDialer.Dial(dcUrl.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer dc.Close()

	go func() {
		defer close(done)
		for {
			// TODO - Data Channel
			_, message, err := dc.ReadMessage()
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
			err := dc.WriteMessage(websocket.TextMessage, buffer)
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
			err := cc.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}

			err = dc.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
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