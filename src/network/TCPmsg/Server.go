package TCPmsg

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"reflect"
	"../../types"
)

type TestMSG struct {
	Number  int    `json:"number"`
	Message string `json:"message"`
}

type HelloMsg struct {
	Message string
	Iter    int
}

type Server struct {
	rxChannels types.RXChannels
	Reader     *bufio.Reader
	Encoder    *json.Encoder
}

type Request struct {
	ChannelAdress string `json:"mAdd"`
	ElevatorId    string `json:"reqId"`
	Data          []byte `json:"data"`
}

type Message struct {
	Connection net.Conn
	Data       []byte
}

func NewServer(rxChs types.RXChannels) *Server {
	server := Server{
		rxChannels: rxChs,
	}
	return &server
}

func (s Server) ListenAndServe(port int, portCh chan<- int) {

	allClients := make(map[net.Conn]string) //map of all clients keyed on their connection
	newConnections := make(chan net.Conn)   //channel for incoming connections
	deadConnections := make(chan net.Conn)  //channel for dead connections
	messages := make(chan Message)          //channel for messages
	var server net.Listener

	//Iterate until free TCP port is found, send port back through channel
	for {
		var err error
		addr := fmt.Sprintf("0.0.0.0:%d", port)
		server, err = net.Listen("tcp", addr)
		if err != nil {
			fmt.Println("Listen err ", err)
			port++
		} else {
			portCh <- port
			break
		}
	}
	log.Printf("TCP Server listening on %s\n", port)

	go acceptConnections(server, newConnections)

	for {
		select {
		case conn := <-newConnections:
			addr := conn.RemoteAddr().String()
			fmt.Printf("Accepted new client, %v\n", addr)
			allClients[conn] = addr
			go read(conn, messages, deadConnections)
		case conn := <-deadConnections:
			fmt.Printf("Client %v disconnected", allClients[conn])
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
		fmt.Println("Error decoding JSON:" + err.Error())
	}
	w := reflect.TypeOf(server.rxChannels)
	x := reflect.ValueOf(server.rxChannels)

	for i := 0; i < w.NumField(); i++ {
		ch := w.Field(i)
		chValue := x.Field(i).Interface()
		T := reflect.TypeOf(chValue).Elem()
		typeName := ch.Tag.Get("addr")
		fmt.Println("TCP: Got type: ", typeName)
		if request.ChannelAdress == typeName {
			v := reflect.New(T)
			err := json.Unmarshal(request.Data, v.Interface())
			if err != nil {
				fmt.Println("Error decoding JSON:" + err.Error())
			}
			reflect.Select([]reflect.SelectCase{{
				Dir:  reflect.SelectSend,
				Chan: reflect.ValueOf(chValue),
				Send: reflect.Indirect(v),
			}})
		}
	}
}
