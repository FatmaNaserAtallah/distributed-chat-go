package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

type Client struct {
	id   int
	ch   chan string // channel used by server to send messages to this client
	conn net.Conn
}

type Server struct {
	mu      sync.Mutex
	clients map[int]*Client
	history []string
	nextID  int
}

func NewServer() *Server {
	return &Server{
		clients: make(map[int]*Client),
		history: make([]string, 0),
		nextID:  1,
	}
}

// broadcast message to all clients except maybe the sender (senderID==0 => broadcast to all)
func (s *Server) broadcast(msg string, senderID int) {
	s.mu.Lock()
	// append to history for every broadcast (including join/leave). adjust as required.
	s.history = append(s.history, msg)
	for id, c := range s.clients {
		if id == senderID { // no self-echo
			continue
		}
		// non-blocking send to avoid goroutine leak if client channel is full
		select {
		case c.ch <- msg:
		default:
			// if the client's channel is full, drop message (or handle disconnect)
		}
	}
	s.mu.Unlock()
}

func (s *Server) handleConnection(conn net.Conn) {
	// assign ID and client struct
	s.mu.Lock()
	id := s.nextID
	s.nextID++
	client := &Client{
		id:   id,
		ch:   make(chan string, 16), // buffered channel
		conn: conn,
	}
	s.clients[id] = client
	// copy history to send to this new client
	historySnapshot := make([]string, len(s.history))
	copy(historySnapshot, s.history)
	s.mu.Unlock()

	// Send existing history to the new client
	go func() {
		writer := bufio.NewWriter(conn)
		for _, line := range historySnapshot {
			fmt.Fprintln(writer, line)
		}
		writer.Flush()
	}()

	// Notify others that this user joined
	joinMsg := fmt.Sprintf("User [%d] joined", id)
	s.broadcast(joinMsg, 0) // senderID=0 => send to all (including new? we'll skip only when sender==id)

	// Goroutine: send messages from server -> client
	go func() {
		w := bufio.NewWriter(conn)
		for msg := range client.ch {
			fmt.Fprintln(w, msg)
			w.Flush()
		}
	}()

	// Goroutine: read lines from client
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		// format message (include sender id or keep as raw)
		msg := fmt.Sprintf("User [%d]: %s", id, text)
		// broadcast to others (skip self)
		s.broadcast(msg, id)
	}
	// client disconnected (scanner ended)
	conn.Close()
	// cleanup
	s.mu.Lock()
	delete(s.clients, id)
	close(client.ch)
	s.mu.Unlock()
	leaveMsg := fmt.Sprintf("User [%d] left", id)
	s.broadcast(leaveMsg, 0)
}

func (s *Server) Run(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	fmt.Println("Server listening on", addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		go s.handleConnection(conn) // concurrent handler
	}
}

func main() {
	server := NewServer()
	if err := server.Run(":9000"); err != nil {
		fmt.Println("Server error:", err)
	}
}
