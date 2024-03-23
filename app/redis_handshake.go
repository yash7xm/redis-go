package main

import (
	"fmt"
	"net"
	"os"
)

func (s *Server) handleHandShake() {
	masterConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", s.replicaOfHost, s.replicaOfPort))
	if masterConn != nil {
		fmt.Printf("Connected to master on %s:%s\n", s.replicaOfHost, s.replicaOfPort)
	}
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	sendPingToMaster(masterConn)
}

func sendPingToMaster(masterConn net.Conn) {
	// sending ping to master
	_, err := masterConn.Write([]byte(GenBulkArray([]string{"PING"})))
	if err != nil {
		fmt.Println(err)
		return
	}

	tempResponse := make([]byte, 1024)
	n, _ := masterConn.Read(tempResponse)
	fmt.Println(string(tempResponse[:n]))

	sendReplConf(masterConn)
}

func sendReplConf(masterConn net.Conn) {
	// sending first replconf to master
	_, err := masterConn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n6380\r\n"))
	if err != nil {
		fmt.Println(err)
		return
	}

	tempResponse := make([]byte, 1024)
	n, _ := masterConn.Read(tempResponse)
	fmt.Println(string(tempResponse[:n]))

	// sending second replconf to master
	_, err = masterConn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"))
	if err != nil {
		fmt.Println(err)
		return
	}

	tempResponse = make([]byte, 1024)
	n, _ = masterConn.Read(tempResponse)
	fmt.Println(string(tempResponse[:n]))

	sendPsyncToMaster(masterConn)
}

func sendPsyncToMaster(masterConn net.Conn) {
	_, err := masterConn.Write([]byte("*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"))
	if err != nil {
		fmt.Println(err)
		return
	}

	tempResponse := make([]byte, 1024)
	n, _ := masterConn.Read(tempResponse)
	fmt.Println(string(tempResponse[:n]))

	tempResponse = make([]byte, 1024)
	n, _ = masterConn.Read(tempResponse)
	fmt.Println(string(tempResponse[:n]))

}

