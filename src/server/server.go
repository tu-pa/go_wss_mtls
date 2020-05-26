package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	//"io"
	"io/ioutil"
	//os"
	//"os/signal"
	"log"
	//"net"
	"crypto/tls"
	"crypto/x509"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"../common"
)

/////////////////////////////////////

type Client struct {
	Id     string
	Joined string
	Conn   *websocket.Conn
	Pool   *Pool
}

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[string]*Client
	Broadcast  chan *common.Request
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan *common.Request),
	}
}

func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			t := time.Now()
			client.Joined = t.Format(time.StampMilli)
			pool.Clients[client.Id] = client
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			for client, _ := range pool.Clients {
				fmt.Println(client)
				//client.Conn.WriteJSON(Message{Type: 1, Body: "New User Joined..."})
			}
			break
		case client := <-pool.Unregister:
			delete(pool.Clients, client.Id)
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			for client, _ := range pool.Clients {
				fmt.Println(client)
				//client.Conn.WriteJSON(Message{Type: 1, Body: "User Disconnected..."})
			}
			break
		case message := <-pool.Broadcast:
			fmt.Println("Sending message to all clients in Pool")
			for _, client := range pool.Clients {
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
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var done = make(chan struct{})
var respCh = make(chan string)

func connect(pool *Pool, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	// Create ID from Remote Address
	var s = strings.ReplaceAll(r.RemoteAddr, ".", "")
	s = strings.ReplaceAll(s, ":", "")

	client := &Client{
		Id:   s,
		Conn: conn,
		Pool: pool,
	}

	pool.Register <- client

	defer func() {
		client.Pool.Unregister <- client
		client.Conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("client response: %s", message)
		respCh <- string(message)
	}
}

func data(pool *Pool, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	// Create ID from Remote Address
	var s = strings.ReplaceAll(r.RemoteAddr, ".", "")
	s = strings.ReplaceAll(s, ":", "")

	client := &Client{
		Id:   s,
		Conn: conn,
		Pool: pool,
	}

	pool.Register <- client

	defer func() {
		client.Pool.Unregister <- client
		client.Conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("TODO: Route Client data: %s", message)
	}
}

func startService(client *Client, w http.ResponseWriter, r *http.Request) {
	if client == nil {
		resp := common.Response{
			ReqId:  "startService",
			Result: "invalid client error",
		}

		log.Println("invalid client: ", resp)

		json.NewEncoder(w).Encode(resp)
		return
	}

	log.Println("Request startService Client: ", client)

	msg, _ := json.Marshal(common.Request{ReqId: common.Reqid_start_service})
	err := client.Conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Println("write:", err)
		return
	}

	// Wait for Response
	select {
	case message := <-respCh:
		resp := &common.Response{
			ReqId:  "",
			Result: "",
		}
		json.Unmarshal([]byte(message), resp)

		json.NewEncoder(w).Encode(resp)
	}
}

func stopService(client *Client, w http.ResponseWriter, r *http.Request) {
	if client == nil {
		resp := common.Response{
			ReqId:  "stopService",
			Result: "invalid client error",
		}

		log.Println("invalid client: ", resp)

		json.NewEncoder(w).Encode(resp)
		return
	}

	log.Println("Request stopService Client: ", client)

	msg, _ := json.Marshal(common.Request{ReqId: common.Reqid_stop_service})
	err := client.Conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Println("write:", err)
		return
	}

	// Wait for Response
	select {
	case message := <-respCh:
		resp := &common.Response{
			ReqId:  "",
			Result: "",
		}
		json.Unmarshal([]byte(message), resp)

		json.NewEncoder(w).Encode(resp)
	}
}

func getAvailClients(pool *Pool, w http.ResponseWriter, r *http.Request) {

	m := []struct {
		Id     string
		Joined string
	}{}

	for client := range pool.Clients {
		m = append(m, struct {
			Id     string
			Joined string
		}{Id: client, Joined: pool.Clients[client].Joined})
	}

	log.Println("Clients: ", m)

	json.NewEncoder(w).Encode(m)
}

// secure ports
var controlPort = ":8442"
var dataPort = ":8443"

// non-secure ports
var webserverPort = ":8444"

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctrlPool := NewPool()
	go ctrlPool.Start()
	dataPool := NewPool()
	go dataPool.Start()

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
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	tlsConfig.BuildNameToCertificate()

	// Create a Server instance to listen on port 8443 with the TLS config
	dataServer := &http.Server{
		Addr:      dataPort,
		TLSConfig: tlsConfig,
	}

	// Set up a resource handler
	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		connect(ctrlPool, w, r)
	})

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		connect(dataPool, w, r)
	})

	// Listen to HTTPS connections with the server certificate and wait
	go func() {
		defer close(done)
		log.Fatal(dataServer.ListenAndServeTLS("certs/cert.pem", "certs/key.pem"))
	}()

	// Create a Server instance to listen on port 8443 with the TLS config
	controlServer := &http.Server{
		Addr:      controlPort,
		TLSConfig: tlsConfig,
	}

	// Listen to HTTPS connections with the server certificate and wait
	go func() {
		defer close(done)
		log.Fatal(controlServer.ListenAndServeTLS("certs/cert.pem", "certs/key.pem"))
	}()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/available_clients", func(w http.ResponseWriter, r *http.Request) {
		getAvailClients(ctrlPool, w, r)
	}).Methods("GET")
	router.HandleFunc("/start_service/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		client := ctrlPool.Clients[id]
		startService(client, w, r)
	}).Methods("GET")
	router.HandleFunc("/stop_service/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		client := ctrlPool.Clients[id]
		stopService(client, w, r)
	}).Methods("GET")

	go func() {
		defer close(done)
		log.Fatal(http.ListenAndServe(webserverPort, router))
	}()

	for {
		select {
		case <-done:
			return
		}
	}
}
