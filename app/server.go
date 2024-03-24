package main

import (
	"bufio"
	"errors"
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
	connectedReplicas ConnectionPool
	replicaMutex      sync.Mutex
}

type ConnectionPool struct {
	replicas []net.Conn
	mutex    sync.Mutex
}

func (cp *ConnectionPool) Add(conn net.Conn) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()
	cp.replicas = append(cp.replicas, conn)
}

// Function to get a connection from the pool
func (cp *ConnectionPool) Get() (net.Conn, error) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()
	if len(cp.replicas) == 0 {
		return nil, errors.New("connection pool is empty")
	}
	conn := cp.replicas[0]
	cp.replicas = cp.replicas[1:]
	return conn, nil
}

// Function to return a connection to the pool
func (cp *ConnectionPool) Put(conn net.Conn) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()
	cp.replicas = append(cp.replicas, conn)
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
	srv.Run(port)

}

func NewServer(role Role, replicaOfHost string, replicaOfPort string) *Server {
	servStorage := NewStorage()

	return &Server{
		storage:           servStorage,
		role:              role,
		replicaOfHost:     replicaOfHost,
		replicaOfPort:     replicaOfPort,
		connectedReplicas: ConnectionPool{},
		replicaMutex:      sync.Mutex{},
	}
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
		s.handleHandShake()
	}

	for {
		conn, err := listener.Accept()
		noOfClients++
		fmt.Println("No of connected clients:- ", noOfClients)
		if err != nil {
			fmt.Println("Error accepting connection try again: ", err.Error())
			os.Exit(1)
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleReplicaPropagation(replicaChannel chan []Value) {
	for {
		args := <-replicaChannel
		go s.propagateSetToReplica(args)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	var replicaChannel chan []Value
	for {
		value, err := DecodeRESP(bufio.NewReader(conn))

		if err != nil {
			fmt.Println("Error decoding RESP: ", err.Error())
			return
		}

		replicaChannel = make(chan []Value)

		if s.role == MasterRole {
			go s.handleReplicaPropagation(replicaChannel)
		}

		s.HandleCommands(value, conn, replicaChannel)
	}
}
