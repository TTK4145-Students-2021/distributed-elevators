package network

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"reflect"

	kcp "gopkg.in/xtaci/kcp-go.v5"
)

type RequestMsg struct {
	ChannelAdress string `json:"mAdd"`
	ElevatorId    string `json:"reqId"`
	Data          []byte `json:"data"`
}

type TcpMsg struct {
	Connection net.Conn
	Data       []byte
}

func listenAndServe(port int, portCh chan<- int, rxChannels RXChannels) {
	allClients := make(map[net.Conn]string) //map of all clients keyed on their connection
	newConnections := make(chan net.Conn)   //channel for incoming connections
	deadConnections := make(chan net.Conn)  //channel for dead connections
	messages := make(chan TcpMsg)           //channel for messages
	var server net.Listener

	//Iterate until free TCP port is found, send port back through channel
	for {
		var err error
		addr := fmt.Sprintf("0.0.0.0:%d", port)
		server, err = kcp.Listen(addr)
		if err != nil {
			fmt.Println("Listen err ", err)
			port++
		} else {
			portCh <- port
			break
		}
	}
	log.Printf("TCP Server listening on %d\n", port)

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
			go PassMsgOnRxChannel(message, rxChannels)
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

func read(conn net.Conn, messages chan TcpMsg, deadConnections chan net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		incoming, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		messages <- TcpMsg{conn, []byte(incoming)}
	}
	deadConnections <- conn
}

func PassMsgOnRxChannel(msg interface{}, rxChannels RXChannels) {
	request := RequestMsg{}
	switch msg := msg.(type) {
	case TcpMsg:
		if err := json.Unmarshal(msg.Data, &request); err != nil {
			fmt.Println("Error decoding JSON:" + err.Error())
		}
	case RequestMsg:
		request = msg
	default:
		fmt.Printf("TCP Server cant handle message type %T", msg)
	}

	w := reflect.TypeOf(rxChannels)
	x := reflect.ValueOf(rxChannels)

	for i := 0; i < w.NumField(); i++ {
		ch := w.Field(i)
		chValue := x.Field(i).Interface()
		T := reflect.TypeOf(chValue).Elem()
		typeName := ch.Tag.Get("addr")
		// fmt.Println("TCP: Got type: ", typeName)
		if request.ChannelAdress == typeName {
			v := reflect.New(T)
			//fmt.Println("Request: ", request)
			err := json.Unmarshal(request.Data, v.Interface())
			if err != nil {
				fmt.Println("Error decoding JSON 2:" + err.Error())
			}
			reflect.Select([]reflect.SelectCase{{
				Dir:  reflect.SelectSend,
				Chan: reflect.ValueOf(chValue),
				Send: reflect.Indirect(v),
			}})
		}
	}
}
