package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type Request struct {
	Id     string `json:"reqId"`
	Action string `json:"action"`
	Data   Data   `json:"data"`
}

type Data struct {
	SomeData string `json:"someData"`
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

	dat := Data{SomeData: "Data boiiii"}

	req := Request{
		Id:     "101",
		Action: "msg",
		Data:   dat,
	}

	bytes, err := json.Marshal(req)
	for {
		_, err = conn.Write(bytes)
		_, err = conn.Write([]byte("\n"))
		if err != nil {
			println("Write to server failed:", err.Error())
		}

		println("write to server = ", string(bytes))

		reply := make([]byte, 1024)

		_, err = conn.Read(reply)
		if err != nil {
			println("Read from server failed:", err.Error())
		}

		println("reply from server=", string(reply))
		time.Sleep(time.Second)
	}

}
