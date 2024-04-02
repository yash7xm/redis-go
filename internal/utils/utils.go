package utils

import (
    "errors"
    "net"
    "sync"
)

type ConnectionPool struct {
    Replicas []net.Conn
    Mutex    sync.Mutex 
}

func (cp *ConnectionPool) Add(conn net.Conn) {
    cp.Mutex.Lock() 
    defer cp.Mutex.Unlock() 
    cp.Replicas = append(cp.Replicas, conn) 
}

// Function to get a connection from the pool
func (cp *ConnectionPool) Get() (net.Conn, error) {
    cp.Mutex.Lock() 
    defer cp.Mutex.Unlock() 
    if len(cp.Replicas) == 0 { 
        return nil, errors.New("connection pool is empty")
    }
    conn := cp.Replicas[0] 
    cp.Replicas = cp.Replicas[1:]
    return conn, nil
}

// Function to return a connection to the pool
func (cp *ConnectionPool) Put(conn net.Conn) {
    cp.Mutex.Lock() 
    defer cp.Mutex.Unlock()
    cp.Replicas = append(cp.Replicas, conn) 
}
