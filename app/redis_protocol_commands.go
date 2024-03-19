package main

import "fmt"

func GenSimpleString(data string) string {
	return fmt.Sprintf("+%s\r\n", data)
}

func GenBulkString(data string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(data), data)
}
