package main

import (
	"fmt"
	"net"
	"strconv"
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

func HandleSetCommand(conn net.Conn, args []Value, storage *Storage) {
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

func HandleReplconfCommand(conn net.Conn) {
	output := GenSimpleString("OK")
	conn.Write([]byte(output))
}

func HandleCommands(value Value, conn net.Conn, storage *Storage, s *Server) {
	command := value.Array()[0].String()
	args := value.Array()[1:]

	switch command {
	case "ping":
		HandlePingCommand(conn)
	case "echo":
		HandleEchoCommand(conn, args[0].String())
	case "set":
		HandleSetCommand(conn, args, storage)
	case "get":
		HandleGetCommand(conn, args[0].String(), storage)
	case "info":
		HandleInfoCommand(conn, s.role)
	case "replconf":
		HandleReplconfCommand(conn)
	default:
		conn.Write([]byte("-ERR unknown command '" + command + "'\r\n"))
	}
}
