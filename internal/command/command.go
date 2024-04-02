package command

import (
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/config"
	"github.com/codecrafters-io/redis-starter-go/internal/parser"
)

func HandlePingCommand(conn net.Conn, s *config.Server) {

	if s.Role == config.MasterRole {
		output := parser.SerializeSimpleString("PONG")
		conn.Write([]byte(output))
	}
}

func HandleEchoCommand(conn net.Conn, value string) {
	output := parser.SerializeBulkString(value)
	conn.Write([]byte(output))
}

func HandleSetCommand(conn net.Conn, args []string, s *config.Server) {
	if len(args) > 2 {
		if args[2] == "px" {
			expiryStr := args[3]
			expiryInMilliSecond, err := strconv.Atoi(expiryStr)
			if err != nil {
				conn.Write([]byte(fmt.Sprintf("-ERR PX value (%s) is not an integer\r\n", expiryStr)))
				return
			}

			s.Storage.SetWithExpiry(args[0], args[1], time.Duration(expiryInMilliSecond)*time.Millisecond)
		}
	} else {
		s.Storage.Set(args[0], args[1])
	}

	if s.Role == config.MasterRole {
		propagateSetToReplica(args, s)
	}

	if s.Role == config.MasterRole {
		conn.Write([]byte("+OK\r\n"))
	}
}

func HandleGetCommand(conn net.Conn, key string, s *config.Server) {
	value, found := s.Storage.Get(key)
	if found {
		output := parser.SerializeBulkString(value)
		conn.Write([]byte(output))
	} else {
		conn.Write([]byte("$-1\r\n"))
	}
}

func HandleInfoCommand(conn net.Conn, s *config.Server) {
	role := s.Role
	if role == "slave" {
		output := parser.SerializeSimpleString("role:slave")
		conn.Write([]byte(output))
	} else if role == "master" {
		replId := "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
		replOffset := 0
		output := parser.SerializeBulkString(fmt.Sprintf("role:%s\r\nmaster_replid:%s\r\nmaster_repl_offset:%d",
			role, replId, replOffset))

		conn.Write([]byte(output))
	}
}

func HandleReplconfCommand(conn net.Conn, args []string, s *config.Server) {
	if strings.ToLower(string(args[0])) == "getack" {
		if s.Role == config.MasterRole {
			fmt.Println(parser.SerializeSimpleError("only slave can receive GETACK"))
		}
		offset := fmt.Sprintf("%d", s.MasterReplOffset)
		response := parser.SerializeArray([]string{"REPLCONF", "ACK", offset})
		conn.Write(response)
		// conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$3\r\nACK\r\n$1\r\n0\r\n"))
	} else {
		output := parser.SerializeSimpleString("OK")
		conn.Write([]byte(output))
	}
}

func HandlePsyncCommand(conn net.Conn, s *config.Server) {
	replId := "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	replOffset := 0

	output := fmt.Sprintf("+FULLRESYNC %s %d\r\n", replId, replOffset)

	s.ReplicaMutex.Lock()
	s.ConnectedReplicas.Add(conn) // Add the replica's connection to the pool
	s.ReplicaMutex.Unlock()

	conn.Write([]byte(output))

	time.Sleep(500 * time.Millisecond)

	sendRdbContent(conn)
}

func sendRdbContent(conn net.Conn) {
	emptyRDBFileBase64 := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
	decodedBytes, err := base64.StdEncoding.DecodeString(emptyRDBFileBase64)

	if err != nil {
		fmt.Println("Error while sendig RDB file", err)
		return
	}

	output := parser.SerializeRDBFileContent(decodedBytes)
	_, err = conn.Write(output)

	if err != nil {
		fmt.Println("Not able to send the RDB file", err)
		return
	}
}

func propagateSetToReplica(args []string, s *config.Server) {

	args = append([]string{"set"}, args...)

	setCommands := parser.SerializeArray(args)

	s.ReplicaMutex.Lock()
	defer s.ReplicaMutex.Unlock()

	// Track the number of successful writes
	successfulWrites := 0

	for {
		replicaConn, err := s.ConnectedReplicas.Get() // Get a connection from the pool
		if err != nil {
			fmt.Println("Error getting connection from pool:", err)
			break // Break loop if there are no available connections
		}

		_, err = replicaConn.Write(setCommands)
		if err != nil {
			fmt.Println("Error writing to replica:", err)
			s.ConnectedReplicas.Put(replicaConn) // Return the connection to the pool
			break
		}

		// Increment successful writes
		successfulWrites++

		// Return the connection to the pool
		s.ConnectedReplicas.Put(replicaConn)

		// Check if all replicas received the command
		if successfulWrites == len(s.ConnectedReplicas.Replicas) {
			break
		}
	}
}

func HandleFullResync(conn net.Conn) {
	conn.Write([]byte("+OK\r\n"))
}

func Handler(value []string, conn net.Conn, s *config.Server) {
	command := strings.ToLower(value[0])
	args := value[1:]

	switch command {
	case "ping":
		HandlePingCommand(conn, s)
	case "echo":
		HandleEchoCommand(conn, args[0])
	case "set":
		HandleSetCommand(conn, args, s)
	case "get":
		HandleGetCommand(conn, args[0], s)
	case "info":
		HandleInfoCommand(conn, s)
	case "replconf":
		HandleReplconfCommand(conn, args, s)
	case "psync":
		HandlePsyncCommand(conn, s)
	case "fullresync":
		HandleFullResync(conn)
	default:
		conn.Write([]byte("-ERR unknown command '" + command + "'\r\n"))
	}
}
