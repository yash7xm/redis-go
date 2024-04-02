package config

import (
	"net"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/internal/utils"
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
	connectedReplicas utils.ConnectionPool
	replicaMutex      sync.Mutex
	masterConn        net.Conn
}

func NewServer(role Role, replicaOfHost string, replicaOfPort string) *Server {
	servStorage := NewStorage()

	return &Server{
		storage:           servStorage,
		role:              role,
		replicaOfHost:     replicaOfHost,
		replicaOfPort:     replicaOfPort,
		connectedReplicas: utils.ConnectionPool{},
		replicaMutex:      sync.Mutex{},
	}
}