
# Simple Chatroom (Go + RPC)

![go](https://img.shields.io/badge/go-1.20-blue) ![status](https://img.shields.io/badge/status-development-yellow) ![docker-pulls](https://img.shields.io/docker/pulls/moazloaie/simple_chatroom?logo=docker)

A simple chatroom written in Go using the native `net/rpc` package for efficient client-server communication. The implementation keeps the latest 1000 messages in memory and enforces per-client history access (clients can only request their own messages).

## Why net/rpc?

Go's native RPC package provides type-safe, efficient communication between Go programs. It offers excellent performance, built-in concurrent handling of connections, and seamless integration with Go's type system.

## Key Components

- 📦 `pkg/chat` — core types and the `ChatService` implementation (`pkg/chat/api.go`). Thread-safe in-memory storage with automatic cleanup (keeps last 1000 messages).
- 🖥️ `cmd/server` — RPC server implementation using Go's net/rpc package for efficient communication (`cmd/server/main.go`).
- 🧑‍💻 `cmd/client` — interactive terminal client that connects via RPC and supports simple commands (`cmd/client/main.go`).

## Quick Start (PowerShell)

### Server

1. Build the server

    ```powershell
    go build ./cmd/server
    ```

2. Run the server (default port 8080)

    ```powershell
    .\server.exe -port 8080
    ```

The server supports graceful shutdown via Ctrl+C.

Server flags:

- **`-port <n>`** — port to listen on (default `8080`)
- **`-logfile <path>`** — write logs to file instead of stdout

### Client

1. Build the client

    ```powershell
    go build ./cmd/client
    ```

2. Run the client and follow the prompt

    ```powershell
    .\client.exe -server localhost:8080
    ```

## Usage

### Client Commands

- **`/help`** — show help and available commands
- **`/history [sinceID]`** — fetch your own messages. If `sinceID` omitted or `0`, returns all messages for your user.
- **`/whoami`** — print current username
- **`/exit`** — quit the client
- Any other text — sent as a message from your username

### Server RPC Interface

The server exposes two RPC methods:

#### ChatService.SendMessage

- **Args:** `SendMessageArgs{ Author string, Text string }`
- **Reply:** `SendMessageReply{ OK bool, Error string }`

#### ChatService.GetHistory

- **Args:** `GetHistoryArgs{ Author string, SinceID int64 }`
- **Reply:** `GetHistoryReply{ Messages []Message }`

where `Message` is:

```go
type Message struct {
    ID        int64     // unique message id
    Author    string    // username
    Text      string    // message body
    Timestamp time.Time // server timestamp
}
```

## Notes & Limitations

- **In-memory storage:** Restarting the server clears the history.
- **No authentication:** The server trusts the `Author` value provided by clients. For real deployments add authentication and server-side identity verification.
- **CLI input cap:** The CLI enforces an ~8KB input cap per message; the server itself does not limit message size today.

## Project Files

- **`pkg/chat/api.go`** — core types and `ChatService`
- **`cmd/server/main.go`** — RPC server implementation
- **`cmd/client/main.go`** — interactive client

## Docker ([Docker Hub Link](https://hub.docker.com/repository/docker/moazloaie/simple_chatroom/general))

This project includes Dockerfiles and a `docker-compose.yml` to build and run the server and client containers locally.

### Images

- **Server image:** `moazloaie/simple_chatroom:server`
- **Client image:** `moazloaie/simple_chatroom:client`

### Build (PowerShell)

Run these from the repository root so the Dockerfiles can access top-level files like `go.mod`:

```powershell
docker build -t moazloaie/simple_chatroom:server -f cmd/server/Dockerfile .
docker build -t moazloaie/simple_chatroom:client -f cmd/client/Dockerfile .
```

### Run with Docker (single containers)

Start the server and publish port 8080 to the host:

```powershell
docker run --rm -d -p 8080:8080 --name simple_chatroom-server moazloaie/simple_chatroom:server
```

Run the client interactively and point it to the host server (Docker Desktop on Windows exposes `host.docker.internal`):

```powershell
docker run --rm -it --name simple_chatroom-client moazloaie/simple_chatroom:client -server host.docker.internal:8080
```

### Run with Docker Compose (recommended for local dev)

The compose file defines a `server` service (background) and a `client` service in the `client` profile (interactive). Recommended workflow:

1. Start the server only (detached):

    ```powershell
    docker compose up -d server
    ```

2. Run the client interactively (attached to the compose network):

    ```powershell
    docker compose run --rm client -server server:8080
    ```

Notes:

- Running `docker compose up` without arguments will start services in your compose file. The `client` service is in a profile and will not start unless explicitly included or run with `docker compose run`.
- If you want both services started automatically (not recommended for most dev workflows because the client is interactive), remove the `client` profile and run `docker compose up`.
- To persist server logs or data, mount a host directory into the server container and pass `-logfile`:

```powershell
docker run --rm -d -p 8080:8080 -v ${PWD}\data:/data --name simple_chatroom-server moazloaie/simple_chatroom:server -logfile /data/server.log
```

## Future Improvements

- ✅ **Add unit tests** for `pkg/chat` (recommended)
- 🔐 **Add authentication** so clients can't impersonate each other
- 💾 **Add persistence** so history survives restarts

## License

This project is licensed under the MIT License.

```text
MIT License

Copyright (c) 2025

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
