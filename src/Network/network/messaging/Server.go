package jsonpipe

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
)
type TestMSG struct {
	number int
	message string
}
type RXChannels struct {
	/*StateCh chan //elev.state
	OrderUpdateCH chan //OrderUpdateCH
	AllOrdersCH chan //orders */
	TestCH chan TestMSG
}
type Server struct {
	rxChannels 	RXChannels
	Reader         *bufio.Reader
	Encoder        *json.Encoder
}

type Request struct {
	ModuleAdress string                 `json:"mAdd"`
	RequestId    string                 `json:"reqId"`
	Data         map[string]interface{} `json:"data"`
}

type Message struct {
	Connection net.Conn
	Data       []byte
}

func NewServer() *Server {
	server := Server{
		ActionRegistry: make(map[string]Action),
	}
	return &server
}

func (s Server) ListenAndServe(port string, busy chan<- bool) {

	allClients := make(map[net.Conn]string) //map of all clients keyed on their connection
	newConnections := make(chan net.Conn)   //channel for incoming connections
	deadConnections := make(chan net.Conn)  //channel for dead connections
	messages := make(chan Message)          //channel for messages

	server, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Listen err ", err)
		busy <- true
		return
	} else {
		busy <- false
	}

	log.Printf("JSON Pipe Server listening on %s\n", port)

	go acceptConnections(server, newConnections)

	for {
		select {
		case conn := <-newConnections:
			addr := conn.RemoteAddr().String()
			log.Printf("Accepted new client, %v", addr)
			allClients[conn] = addr
			go read(conn, messages, deadConnections)
		case conn := <-deadConnections:
			log.Printf("Client %v disconnected", allClients[conn])
			delete(allClients, conn)
		case message := <-messages:
			go s.HandleMessage(message)
		}
	}

}

func acceptConnections(server net.Listener, newConnections chan net.Conn) {
	for {
		conn, err := server.Accept()
		if err != nil {
			fmt.Println(err)
		}
		newConnections <- conn
	}
}

func read(conn net.Conn, messages chan Message, deadConnections chan net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		incoming, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		messages <- Message{conn, []byte(incoming)}
	}
	deadConnections <- conn
}

func (server Server) HandleMessage(msg Message) {

	request := Request{}

	if err := json.Unmarshal(msg.Data, &request); err != nil {
		log.Println("Error decoding JSON:" + err.Error())
	}

	switch request.ModuleAdress{
		case "master"
			masterCh<-request.Data
		case "orders"
			ordersCh<-request.Data
		case "peer"
			peerCh<-request.Data
	}

	/*bytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling JSON:%s\n", err)
		return
	}

	msg.Connection.Write(bytes)*/

	return
}
