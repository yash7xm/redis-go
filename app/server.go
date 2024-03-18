package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

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

		LogConnection(conn)
		RespondWith(conn, respPong)
	}
}

func LogConnection(conn net.Conn) {
	str, err := ReadOutConnection(conn)

	if err != nil {
		fmt.Printf("Failed to read connection: %v\n", err)
		return
	}
	fmt.Println("Recieved new request:")
	fmt.Println(str)
}

const respNoContent = "HTTP/1.1 204 No Content\r\n"

const respPong = "+PONG\r\n"

func RespondWith(conn net.Conn, response string) {
	defer conn.Close()

	written, err := conn.Write([]byte(response))

	if err != nil {
		fmt.Printf("Failed to response %v\n", err)
		return
	}

	fmt.Printf("Wrote %v response bytes\n", written)
}

func ReadOutConnection(conn net.Conn) (string, error) {
	// Initialize buffer size and string builder
	buffSize := 1024
	builder := strings.Builder{}
	bytes := make([]byte, buffSize)
	amount := buffSize

	// Read data from the connection until there's no more data
	for amount == buffSize {
		am, err := conn.Read(bytes)
		amount = am
		if err != nil {
			return "", err
		}
		builder.Write(bytes)
	}

	// Return the complete string and nil error
	return builder.String(), nil
}
