package main

import (
	"errors"
	"net"
	"sync"
)

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
