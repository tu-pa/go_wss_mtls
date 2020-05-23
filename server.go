package main

import (
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"log"
	"net"
	"net/http"
	"crypto/x509"
	"crypto/tls"

	"github.com/gorilla/websocket"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Write "Hello, world!" to the response body
	io.WriteString(w, "Hello, world!\n")
}

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	ip, port, err := net.SplitHostPort(r.RemoteAddr)
	log.Print("connection from: ",ip,":",port)

	defer c.Close()
	for {
		// Echo Loop
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

var controlPort = ":8442"
var dataPort = ":8443"

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Set up a /hello resource handler
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/echo", echo)

	// Create a CA certificate pool and add cert.pem to it
	caCert, err := ioutil.ReadFile("certs/cert.pem")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create the TLS Config with the CA pool and enable Client certificate validation
	tlsConfig := &tls.Config{
		ClientCAs: caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	tlsConfig.BuildNameToCertificate()

	// Create a Server instance to listen on port 8443 with the TLS config
	dataServer := &http.Server{
		Addr:      dataPort,
		TLSConfig: tlsConfig,
	}

	done := make(chan struct{})

	// Listen to HTTPS connections with the server certificate and wait
	go func(){
		defer close(done)
		log.Fatal(dataServer.ListenAndServeTLS("certs/cert.pem", "certs/key.pem"))
	}()

	// Create a Server instance to listen on port 8443 with the TLS config
	controlServer := &http.Server{
		Addr:      controlPort,
		TLSConfig: tlsConfig,
	}

	// Listen to HTTPS connections with the server certificate and wait
	go func(){
		defer close(done)
		log.Fatal(controlServer.ListenAndServeTLS("certs/cert.pem", "certs/key.pem"))
	}()

	// Wait loop
	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			select {
			case <-done:
			}
			return
		}
	}
}