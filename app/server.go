package main

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/config"
	"github.com/codecrafters-io/redis-starter-go/internal/parser"
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

	role := config.MasterRole

	if replicaOfHost != "" {
		role = config.SlaveRole
	}

	srv := config.NewServer(role, replicaOfHost, replicaOfPort)
	Run(port, srv)
}

var noOfClients int

func Run(port string, s *config.Server) {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	fmt.Printf("Server started on %s\n", port)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	if s.Role == config.SlaveRole {
		go handleHandShake(s)
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

		go handleConnection(conn, s)
	}
}

func handleConnection(conn net.Conn, s *config.Server) {
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
