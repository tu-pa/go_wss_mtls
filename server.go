package main

import (
	"fmt"
	//"io"
	"io/ioutil"
	//os"
	//"os/signal"
	"log"
	//"net"
	"net/http"
	"crypto/x509"
	"crypto/tls"

	"github.com/gorilla/websocket"
)

// Control Channel Request/Response //
var reqid_start_service = "start_service"
var reqid_stop_service = "stop_service"
var reqid_get_avail_clients = "get_available_clients"
var reqid_get_client_statistics = "get_client_statistics"
var reqid_get_client_ip_addr = "get_client_ip_addr"
var reqid_start_client_data_test = "start_data_test"
var reqid_stop_client_data_test = "stop_get_data_test"

type Request struct {
	reqId string
	clientId string
}

type Response struct {
	reqId string
	result string
	data string
	clients []string
}
/////////////////////////////////////

type Client struct {
    id   string
    Conn *websocket.Conn
    Pool *Pool
}

type Pool struct {
    Register   chan *Client
    Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan *Request
}

func NewPool() *Pool {
    return &Pool{
        Register:   make(chan *Client),
        Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan *Request),
    }
}

func (pool *Pool) Start() {
    for {
        select {
        case client := <-pool.Register:
            pool.Clients[client] = true
            fmt.Println("Size of Connection Pool: ", len(pool.Clients))
            for client, _ := range pool.Clients {
                fmt.Println(client)
                //client.Conn.WriteJSON(Message{Type: 1, Body: "New User Joined..."})
            }
            break
        case client := <-pool.Unregister:
            delete(pool.Clients, client)
            fmt.Println("Size of Connection Pool: ", len(pool.Clients))
            for client, _ := range pool.Clients {
				fmt.Println(client)
                //client.Conn.WriteJSON(Message{Type: 1, Body: "User Disconnected..."})
            }
			break
		case message := <-pool.Broadcast:
            fmt.Println("Sending message to all clients in Pool")
            for client, _ := range pool.Clients {
                if err := client.Conn.WriteJSON(message); err != nil {
                    fmt.Println(err)
                    return
                }
            }
		}
    }
}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool { return true },
}

func control(pool *Pool, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

    client := &Client{
        Conn: conn,
        Pool: pool,
	}

	pool.Register <- client

	defer func() {
        client.Pool.Unregister <- client
        client.Conn.Close()
	}()

	for {
		// TODO - Control Loop
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = conn.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

var controlPort = ":8442"
var dataPort = ":8443"
var webserverPort = ":8444"

func main() {

	pool := NewPool()
    go pool.Start()

	// Set up a resource handler
	http.HandleFunc("/control", func(w http.ResponseWriter, r *http.Request) {
        control(pool, w, r)
	})

	// Local Webserver Feature List
	// Start Client Service
	// - Client Response Ok/Fail
	// Stop Client Service
	// - Client Response Ok/Fail
	// Get Available Clients
	// - Server Response List of Clients
	// Get Client Statistics
	// - Client Response Data or Fail
	// Get Client IP Address
	// - Client Response with IP Address or Fail
	// Start Client Data Test
	// - Client Response Ok/Fail
	// Stop/Get Client Data Test
	// - Client Response Ok/Fail

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

	webserver := &http.Server{Addr: webserverPort}
	go func(){
		defer close(done)
		log.Fatal(webserver.ListenAndServe())
	}()

    for {
		select {
		case <-done:
			return
		}
	}
}