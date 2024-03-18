package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

const respPong = "+PONG\r\n"

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	_, err := conn.Read(buf)

	if errors.Is(err, io.EOF) {
		return
	}

	if err != nil {
		panic(err)
	}

	conn.Write([]byte(respPong))
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}
