package main

import (
	"fmt"
	"net"
	"os"
)

func handleHandShake(replicaOfHost string, replicaOfPort string) {
	for {
		masterConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", replicaOfHost, replicaOfPort))
		if err != nil {
			fmt.Println("Not able to connect to master", err)
			os.Exit(1)
		}
		go sendPingToMaster(masterConn)
	}
}

func sendPingToMaster(masterConn net.Conn) {
	// sending ping to master
	_, err := masterConn.Write([]byte(GenBulkArray([]string{"PING"})))
	if err != nil {
		fmt.Println(err)
		return
	}
	go sendReplConf(masterConn)
}

func sendReplConf(masterConn net.Conn) {
	// sending first replconf to master
	_, err := masterConn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n6380\r\n"))
	if err != nil {
		fmt.Println(err)
		return
	}

	// sending second replconf to master
	_, err = masterConn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"))
	if err != nil {
		fmt.Println(err)
		return
	}
}
