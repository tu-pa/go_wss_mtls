package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"flag"
	"math/rand"
	"net"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"

	"../common"
)

var controlChannel = flag.String("caddr", "127.0.0.1:8442", "control channel")
var dataChannel = flag.String("daddr", "127.0.0.1:8443", "data channel")
var udpPort = ":8444"

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	//// Secure Websocket Setup: 2-Way Auth ////
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

	wssDialer := websocket.DefaultDialer
	wssDialer.TLSClientConfig = &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	}

	//// Control Channel Setup ////
	ccUrl := url.URL{Scheme: "wss", Host: *controlChannel, Path: "/connect"}
	log.Printf("connecting control channel to %s", ccUrl.String())

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
				log.Println(err)
				return
			}
			log.Printf("recv: %s", message)

			req := &common.Request{
				ReqId: "",
			}

			json.Unmarshal([]byte(message), req)

			// TOOD
			log.Println("TODO: Request: ", req.ReqId)

			resp := &common.Response{
				ReqId:  "",
				Result: "",
			}
			resp.ReqId = req.ReqId

			if req.ReqId == common.Reqid_start_service {
				// TODO: systemctl start clientService ?
				resp.Result = "ok"
			} else if req.ReqId == common.Reqid_stop_service {
				// TODO: systemctl stop clientService ?
				resp.Result = "ok"
			} else {
				resp.Result = "unhandled event"
			}

			respMsg, _ := json.Marshal(resp)

			err = cc.WriteMessage(websocket.TextMessage, respMsg)
			if err != nil {
				log.Println("write:", err)
				return
			}

		}
	}()

	//// Data Channel Setup ////
	dcUrl := url.URL{Scheme: "wss", Host: *dataChannel, Path: "/data"}
	log.Printf("connecting data channel to %s", dcUrl.String())

	dcConn, _, err := wssDialer.Dial(dcUrl.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
		return
	}
	defer dcConn.Close()

	go func() {
		defer close(done)
		for {
			// TODO - Data Channel
			_, message, err := dcConn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("TODO: Data Channel recv: %s", message)
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

	go func() {
		defer udpCon.Close()
		buffer := make([]byte, 1024)
		rand.Seed(time.Now().Unix())

		defer close(done)
		for {
			udpCon.ReadFromUDP(buffer)

			err := dcConn.WriteMessage(websocket.TextMessage, buffer)
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}()

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
				log.Println(err)
				return
			} else {
				log.Println("closed")
			}

			err = dcConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println(err)
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
