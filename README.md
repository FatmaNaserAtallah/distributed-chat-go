# Distributed Chat Server (Go)

This project implements a simple real-time chat server in Go using goroutines and channels.

## Features
- Multiple TCP clients can connect concurrently.
- Server sends the full message history to any newly connected client.
- When a client joins: server broadcasts `User [ID] joined`.
- When a client sends a message: server broadcasts `User [ID]: <message>` to all other clients (no self-echo).
- When a client leaves: server broadcasts `User [ID] left`.
- Shared client list and history are synchronized using `sync.Mutex`.

## Files
- `server.go` — server implementation.
- `client.go` — simple client to connect and send messages.

## How to run (local)
1. Start server: go run server.go
2. Start clients: go run client.go localhost:9000 in separate terminals
3. Type messages in clients to test broadcasting and history.
