package main

import (
	"fmt"
	"net"
	"os"
	"sync"
)

func (s *Server) handleHandShake(wg *sync.WaitGroup) *net.Conn {
	masterConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", s.replicaOfHost, s.replicaOfPort))
	if masterConn != nil {
		fmt.Printf("Connected to master on %s:%s\n", s.replicaOfHost, s.replicaOfPort)
		
		// s.connectedReplicas.Add(masterConn)
	}
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	// sending ping to master
	_, err = masterConn.Write([]byte(GenBulkArray([]string{"PING"})))
	if err != nil {
		fmt.Println(err)
		return nil
	}

	tempResponse := make([]byte, 1024)
	n, _ := masterConn.Read(tempResponse)
	fmt.Println(string(tempResponse[:n]))

	// sending first replconf to master
	_, err = masterConn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n6380\r\n"))
	if err != nil {
		fmt.Println(err)
		return nil
	}

	tempResponse = make([]byte, 1024)
	n, _ = masterConn.Read(tempResponse)
	fmt.Println(string(tempResponse[:n]))

	// sending second replconf to master
	_, err = masterConn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"))
	if err != nil {
		fmt.Println(err)
		return nil
	}

	tempResponse = make([]byte, 1024)
	n, _ = masterConn.Read(tempResponse)
	fmt.Println(string(tempResponse[:n]))

	_, err = masterConn.Write([]byte("*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"))
	if err != nil {
		fmt.Println(err)
		return nil
	}

	tempResponse = make([]byte, 1024)
	n, _ = masterConn.Read(tempResponse)
	fmt.Println(string(tempResponse[:n]))

	tempResponse = make([]byte, 1024)
	n, _ = masterConn.Read(tempResponse)
	fmt.Println(string(tempResponse[:n]))

	wg.Done()

	return &masterConn
}
