package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	storage := NewStorage()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection try again: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn, storage)
	}
}

func handleConnection(conn net.Conn, storage *Storage) {
	defer conn.Close()
	for {
		value, err := DecodeRESP(bufio.NewReader(conn))

		if err != nil {
			fmt.Println("Error decoding RESP: ", err.Error())
			return // Ignore clients that we fail to read from
		}

		command := value.Array()[0].String()
		args := value.Array()[1:]

		switch command {
		case "ping":
			conn.Write([]byte("+PONG\r\n"))
		case "echo":
			conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(args[0].String()), args[0].String())))
		case "set":
			storage.Set(args[0].String(), args[1].String())
			conn.Write([]byte("+OK\r\n"))
		case "get":
			value, found := storage.Get(args[0].String())
			if found {
				conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)))
			} else {
				conn.Write([]byte("-ERR unknown command '" + command + "'\r\n"))
			}
		default:
			conn.Write([]byte("-ERR unknown command '" + command + "'\r\n"))
		}
	}
}
