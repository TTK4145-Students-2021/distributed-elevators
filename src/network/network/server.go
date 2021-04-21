package network

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
)

type netMsg struct {
	ChannelAddress string `json:"mAdd"`
	Data          []byte `json:"data"`
}

type connectionMsg struct {
	Connection net.Conn
	Data       []byte
}

func runServer(port int, connection net.Listener, rxChannels RXChannels) {
	allClients := make(map[net.Conn]string)
	newConnections := make(chan net.Conn)
	deadConnections := make(chan net.Conn)
	messages := make(chan connectionMsg)

	go acceptConnections(connection, newConnections)

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
			go decodeMsg(message, rxChannels)
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

func read(conn net.Conn, messages chan connectionMsg, deadConnections chan net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		incoming, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		messages <- connectionMsg{conn, []byte(incoming)}
	}
	deadConnections <- conn
}

func decodeMsg(msg interface{}, rxChannels RXChannels) {
	request := netMsg{}
	switch msg := msg.(type) {
	case connectionMsg:
		if err := json.Unmarshal(msg.Data, &request); err != nil {
			fmt.Println("Error decoding JSON:" + err.Error())
		}
	case netMsg:
		request = msg
	default:
		fmt.Printf("TCP Server cant handle message type %T", msg)
	}
	//reflect implementation to allow marshaling data to correct struct,
	// then send to correct rxChannel given by addr tag
	w := reflect.TypeOf(rxChannels)
	x := reflect.ValueOf(rxChannels)

	for i := 0; i < w.NumField(); i++ {
		ch := w.Field(i)
		chValue := x.Field(i).Interface()
		T := reflect.TypeOf(chValue).Elem()
		typeName := ch.Tag.Get("addr")
		if request.ChannelAddress == typeName {
			v := reflect.New(T)
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
