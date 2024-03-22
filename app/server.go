package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
)

const (
	MasterRole Role = "master"
	SlaveRole  Role = "slave"
)

type Role string

type Server struct {
	storage           *Storage
	role              Role
	replicaOfHost     string
	replicaOfPort     string
	connectedReplicas []*net.Conn
	replicaMutex      sync.Mutex
	commandQueue      chan []Value
}

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
	go srv.handleCommandPropagation()
	srv.Run(port)

}

func NewServer(role Role, replicaOfHost string, replicaOfPort string) *Server {
	servStorage := NewStorage()

	return &Server{
		storage:       servStorage,
		role:          role,
		replicaOfHost: replicaOfHost,
		replicaOfPort: replicaOfPort,
		commandQueue:  make(chan []Value, 100),
	}
}

func (s *Server) handleCommandPropagation() {
    for {
        select {
        case args := <-s.commandQueue:
            s.propagateToReplicas(args)
        }
    }
}

func (s *Server) propagateToReplicas(args []Value) {
    s.replicaMutex.Lock()
    defer s.replicaMutex.Unlock()

    for i := len(s.connectedReplicas) - 1; i >= 0; i-- {
        conn := s.connectedReplicas[i]

        command := SerializeArray(
            SerializeBulkString("SET"),
            SerializeBulkString(args[0].String()),
            SerializeBulkString(args[1].String()),
        )
        _, err := (*conn).Write([]byte(command))
        if err != nil {
            // Replica disconnected, remove from connected replicas
            fmt.Println("Replica disconnected:", err)
            (*conn).Close() // Close the connection
            s.connectedReplicas = append(s.connectedReplicas[:i], s.connectedReplicas[i+1:]...)
        }
    }
}

func (s *Server) Run(port string) {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	fmt.Printf("Server started on %s\n", port)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	if s.role == SlaveRole {
		handleHandShake(s)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection try again: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn, s)
	}
}

func handleConnection(conn net.Conn, s *Server) {
	defer conn.Close()
	for {
		value, err := DecodeRESP(bufio.NewReader(conn))

		if err != nil {
			fmt.Println("Error decoding RESP: ", err.Error())
			return // Ignore clients that we fail to read from
		}

		HandleCommands(value, conn, s)
	}
}
