package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run client.go <server:port>")
		return
	}
	addr := os.Args[1]
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Dial error:", err)
		return
	}
	defer conn.Close()

	// Goroutine: read from server and print
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		fmt.Println("Server closed connection")
		os.Exit(0)
	}()

	// Main: read from stdin and send to server
	stdinScanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(conn)
	for stdinScanner.Scan() {
		text := stdinScanner.Text()
		if text == "" {
			continue
		}
		fmt.Fprintln(writer, text)
		writer.Flush()
	}
}