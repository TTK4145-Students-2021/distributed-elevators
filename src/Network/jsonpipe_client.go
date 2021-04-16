package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type Request struct {
	/*Id     string `json:"reqId"`
	Action string `json:"action"`
	Data   Data   `json:"data"`*/
	ChannelAdress string `json:"mAdd"`
	ElevatorId    string `json:"reqId"`
	Data          []byte `json:"data"`
}

type TestMSG struct {
	Number  int    `json:"number"`
	Message string `json:"message"`
}

func main() {
	sendSomeData()
}

func sendSomeData() {
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		fmt.Println("Network ")
		fmt.Println(err)
		return
	}
	defer conn.Close()

	fmt.Println("Sending some data")

	data := &TestMSG{42, "Data boiiii"}
	//fmt.Println("Struct: ", data)
	dat, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Data: %s\n", string(dat))
	//data2 := TestMSG{}
	var data2 TestMSG
	if err := json.Unmarshal(dat, &data2); err != nil {
		fmt.Println("Error decoding JSON:" + err.Error())
	}
	fmt.Println("Unmarshal: ", data2)
	req := Request{
		ElevatorId:    "101",
		ChannelAdress: "testch1",
		Data:          dat,
	}

	bytes, err := json.Marshal(req)
	for {
		_, err = conn.Write(bytes)
		_, err = conn.Write([]byte("\n"))
		if err != nil {
			println("Write to server failed:", err.Error())
		}

		println("write to server = ", string(bytes))

		/*_, err = conn.Read(reply)
		if err != nil {
			println("Read from server failed:", err.Error())
		}

		println("reply from server=", string(reply))*/
		time.Sleep(time.Second)
	}

}
