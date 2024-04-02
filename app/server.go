package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/internal/parser"
	"github.com/codecrafters-io/redis-starter-go/internal/command"
)



func main() {
	args := os.Args[1:]

	var replicaOfHost string
	var replicaOfPort string
	port := "6379"

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--replicaof":
			if i+2 < len(args) {
				replicaOfHost = args[i+1]
				replicaOfPort = args[i+2]
			}
		case "--port":
			if i+1 < len(args) {
				port = args[i+1]
			}
		}
	}

	role := MasterRole

	if replicaOfHost != "" {
		role = SlaveRole
	}

	srv := NewServer(role, replicaOfHost, replicaOfPort)
	srv.Run(port)

}


var noOfClients int

func (s *Server) Run(port string) {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	fmt.Printf("Server started on %s\n", port)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	if s.role == SlaveRole {
		go s.handleHandShake()
	}

	for {
		conn, err := listener.Accept()
		noOfClients++
		fmt.Println("No of connected clients:- ", noOfClients)
		fmt.Println("Connected to ", conn.RemoteAddr())
		if err != nil {
			fmt.Println("Error accepting connection try again: ", err.Error())
			os.Exit(1)
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := parser.Deserialize(reader)

		if err != nil {
			fmt.Println("Error decoding RESP: ", err.Error())
			return
		}

		fmt.Println("Commands: ", message.Commands)

		if len(message.Commands) == 0 {
			continue
		}

		command.Handler(message.Commands, conn, s)
	}
}
