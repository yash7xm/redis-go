package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/parser"
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

	// sending ping to master
	_, err = masterConn.Write([]byte(GenBulkArray([]string{"PING"})))
	if err != nil {
		fmt.Println(err)
		return
	}

	tempResponse := make([]byte, 1024)
	n, _ := masterConn.Read(tempResponse)
	fmt.Println(string(tempResponse[:n]))

	// sending first replconf to master
	_, err = masterConn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n6380\r\n"))
	if err != nil {
		fmt.Println(err)
		return
	}

	n, _ = masterConn.Read(tempResponse)
	fmt.Println(string(tempResponse[:n]))

	// sending second replconf to master
	_, err = masterConn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"))
	if err != nil {
		fmt.Println(err)
		return
	}

	n, _ = masterConn.Read(tempResponse)
	fmt.Println(string(tempResponse[:n]))

	// sending psync
	_, err = masterConn.Write([]byte("*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"))
	if err != nil {
		fmt.Println(err)
		return
	}

	reader := bufio.NewReader(masterConn)
	for {
		message, err := parser.Deserialize(reader)
		if err != nil {
			fmt.Println("Error decoding RESP: ", err.Error())
			return
		}

		if strings.ToUpper(message.Commands[0]) == "FULLRESYNC" {
			fmt.Println("Recieved PSYNC response from master")

			err := parser.ExpectRDBFile(reader)
			if err != nil {
				fmt.Println("Error expecting RDB file: ", err.Error())
				break
			}
			fmt.Println("RDB file received")
			continue
		}

		fmt.Println("Commands: ", message.Commands)

		if len(message.Commands) == 0 {
			continue
		}

		s.HandleCommands(message.Commands, masterConn)
	}

	s.masterConn = masterConn
}
