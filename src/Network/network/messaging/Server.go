package jsonpipe

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"reflect"
)

type TestMSG struct {
	Number  int    `json:"number"`
	Message string `json:"message"`
}
type RXChannels struct {
	/*StateCh chan //elev.state
	OrderUpdateCH chan //OrderUpdateCH
	AllOrdersCH chan //orders */
	TestCh1 chan TestMSG `addr:"testch1"`
	TestCh2 chan TestMSG `addr:"testch2"`
}
type Server struct {
	rxChannels RXChannels
	Reader     *bufio.Reader
	Encoder    *json.Encoder
}

type Request struct {
	ChannelAdress string `json:"mAdd"`
	RequestId     string `json:"reqId"`
	//Data          map[string]interface{} `json:"data"`
	Data []byte `json:"data"`
}

type Message struct {
	Connection net.Conn
	Data       []byte
}

func NewServer(rxChs RXChannels) *Server {
	server := Server{
		rxChannels: rxChs,
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
			fmt.Printf("Accepted new client, %v", addr)
			allClients[conn] = addr
			go read(conn, messages, deadConnections)
		case conn := <-deadConnections:
			fmt.Printf("Client %v disconnected", allClients[conn])
			delete(allClients, conn)
		case message := <-messages:
			fmt.Printf("Got message\n")
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
	fmt.Println("Message: ", msg.Data)
	if err := json.Unmarshal(msg.Data, &request); err != nil {
		log.Println("Error decoding JSON:" + err.Error())
	}
	fmt.Println("Request mAdd: ", request.ChannelAdress)
	fmt.Println("Request mAdd: ", request.RequestId)
	fmt.Println("Request Data: ", request.Data)
	w := reflect.TypeOf(server.rxChannels)
	x := reflect.ValueOf(server.rxChannels)

	for i := 0; i < w.NumField(); i++ {
		ch := w.Field(i)
		chV := x.Field(i).Interface()
		//for _, ch := range server.rxChannels {
		fmt.Printf("Addr: %s\n", ch.Tag.Get("addr"))
		//X := reflect.TypeOf(ch).Elem()
		T := reflect.TypeOf(chV).Elem()
		//typeName := T.String()
		typeName := ch.Tag.Get("addr")

		if request.ChannelAdress == typeName {
			fmt.Printf("Typename: %s\n", typeName)
			v := reflect.New(T)
			err := json.Unmarshal(request.Data, v.Interface())
			if err != nil {
				fmt.Println("Error decoding JSON:" + err.Error())
			}
			//fmt.Printf("request Data: %s\n", request.Data)
			//fmt.Printf("Sending on channel: %s\n", T)
			//fmt.Printf("Chan: %s\n", reflect.ValueOf(chV))
			reflect.Select([]reflect.SelectCase{{
				Dir:  reflect.SelectSend,
				Chan: reflect.ValueOf(chV),
				Send: reflect.Indirect(v),
			}})
		}
	}
	/*switch request.ModuleAdress{
		case "TestCh1"
			orderUpdate OrderUpdate{}
			err := mapstructure.Decode(input, &restaurant)
			server.rxChannels.TestCh1<-request.Data
		case "TestCh2"
			server.rxChannels.TestCh2<-request.Data
	}*/

	/*bytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling JSON:%s\n", err)
		return
	}

	msg.Connection.Write(bytes)*/

	return
}
