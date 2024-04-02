package config

import (
	"sync"

	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/utils"
)

const (
	MasterRole Role = "master"
	SlaveRole  Role = "slave"
)

type Role string

type Server struct {
	Storage           *store.Storage
	Role              Role
	ReplicaOfHost     string
	ReplicaOfPort     string
	ConnectedReplicas utils.ConnectionPool
	ReplicaMutex      sync.Mutex
}

func NewServer(role Role, replicaOfHost string, replicaOfPort string) *Server {
	servStorage := store.NewStorage()

	return &Server{
		Storage:           servStorage,
		Role:              role,
		ReplicaOfHost:     replicaOfHost,
		ReplicaOfPort:     replicaOfPort,
		ConnectedReplicas: utils.ConnectionPool{},
		ReplicaMutex:      sync.Mutex{},
	}
}
