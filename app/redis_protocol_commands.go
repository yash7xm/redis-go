package main

import (
	"fmt"
	"strconv"
	"strings"
)

func GenSimpleString(data string) string {
	return fmt.Sprintf("+%s\r\n", data)
}

func GenBulkString(data string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(data), data)
}

func GenBulkArray(input []string) string {
	return fmt.Sprintf("*%d\r\n%s", len(input), createBulkString(input))
}

func createBulkString(input []string) string {
	result := ""
	for _, item := range input {
		result = fmt.Sprintf("%s%s\r\n", result, item)
	}

	return fmt.Sprintf("$%d\r\n%s", len(result)-2, result)
}

func SerializeBulkString(s string) string {
	v := fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
	return v
}

func SerializeNullBulkString() string {
	return "$-1\r\n"
}

func SerializeArray(elements ...string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*%d\r\n", len(elements)))
	for _, str := range elements {
		sb.WriteString(str)
	}
	return sb.String()
}

func RDBFileContent(message []byte) []byte {
	return []byte(fmt.Sprintf("$%s\r\n%s", strconv.Itoa(len(message)), message))
}
