package main

import (
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func HandlePingCommand(conn net.Conn) {
	output := GenSimpleString("PONG")
	conn.Write([]byte(output))
}

func HandleEchoCommand(conn net.Conn, value string) {
	output := GenBulkString(value)
	conn.Write([]byte(output))
}

func HandleSetCommand(conn net.Conn, args []Value, storage *Storage, s *Server) {
	if s.role == MasterRole {
		sendToAllTheReplicas(args, s)
	}
	if len(args) > 2 {
		if args[2].String() == "px" {
			expiryStr := args[3].String()
			expiryInMilliSecond, err := strconv.Atoi(expiryStr)
			if err != nil {
				conn.Write([]byte(fmt.Sprintf("-ERR PX value (%s) is not an integer\r\n", expiryStr)))
				return
			}

			storage.SetWithExpiry(args[0].String(), args[1].String(), time.Duration(expiryInMilliSecond)*time.Millisecond)
		}
	} else {
		storage.Set(args[0].String(), args[1].String())
	}
	conn.Write([]byte("+OK\r\n"))
}

func HandleGetCommand(conn net.Conn, key string, storage *Storage) {
	value, found := storage.Get(key)
	if found {
		output := GenBulkString(value)
		conn.Write([]byte(output))
	} else {
		conn.Write([]byte("$-1\r\n"))
	}
}

func HandleInfoCommand(conn net.Conn, role Role) {
	if role == "slave" {
		output := GenSimpleString("role:slave")
		conn.Write([]byte(output))
	} else if role == "master" {
		replId := "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
		replOffset := 0
		output := GenBulkString(fmt.Sprintf("role:%s\r\nmaster_replid:%s\r\nmaster_repl_offset:%d",
			role, replId, replOffset))

		conn.Write([]byte(output))
	}
}

func HandleReplconfCommand(conn net.Conn, s *Server) {
	output := GenSimpleString("OK")
	conn.Write([]byte(output))

	s.connectedReplicas = append(s.connectedReplicas, &conn)
	fmt.Println(conn)
}

func HandlePsyncCommand(conn net.Conn) {
	replId := "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	replOffset := 0

	output := fmt.Sprintf("+FULLRESYNC %s %d\r\n", replId, replOffset)
	conn.Write([]byte(output))

	sendRdbContent(conn)
}

func sendRdbContent(conn net.Conn) {
	emptyRDBFileBase64 := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
	decodedBytes, err := base64.StdEncoding.DecodeString(emptyRDBFileBase64)

	if err != nil {
		fmt.Println("Error while sendig RDB file", err)
		return
	}

	output := RDBFileContent(decodedBytes)
	_, err = conn.Write(output)

	if err != nil {
		fmt.Println("Not able to send the RDB file", err)
		return
	}
}

func sendToAllTheReplicas(args []Value, s *Server) {
	fmt.Println("Connected Replicas:- ", s.connectedReplicas)
	for _, conn := range s.connectedReplicas {
		fmt.Println(args)
		var arr []string
		arr = append(arr, "set")
		for _, arg := range args {
			arr = append(arr, arg.String())
		}
		output := GenBulkArray(arr)
		fmt.Println("output of set command is :- ", output)
		(*conn).Write([]byte(output))
	}
}

func HandleCommands(value Value, conn net.Conn, storage *Storage, s *Server) {
	command := strings.ToLower(value.Array()[0].String())
	args := value.Array()[1:]

	fmt.Println(s.connectedReplicas)

	switch command {
	case "ping":
		HandlePingCommand(conn)
	case "echo":
		HandleEchoCommand(conn, args[0].String())
	case "set":
		HandleSetCommand(conn, args, storage, s)
	case "get":
		HandleGetCommand(conn, args[0].String(), storage)
	case "info":
		HandleInfoCommand(conn, s.role)
	case "replconf":
		HandleReplconfCommand(conn, s)
	case "psync":
		HandlePsyncCommand(conn)
	default:
		conn.Write([]byte("-ERR unknown command '" + command + "'\r\n"))
	}
}
