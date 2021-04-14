package main

import (
	"fmt"

	jsonpipe "./network/messaging"
)

func main() {
	port := 8080
	handler := MessageHandler()
	server := jsonpipe.NewServer()
	server.Handle("msg", handler)
	for {
		adress := fmt.Sprintf("0.0.0.0:%d", port)
		err := server.ListenAndServe(adress)
		if err > 0 {
			break
		}
		port++
	}
}

func MessageHandler() jsonpipe.Handler {
	return func(response *jsonpipe.Response, request *jsonpipe.Request) {
		fmt.Println("Data: ", request.Data)
		response.Data = "Message received"
	}
}
