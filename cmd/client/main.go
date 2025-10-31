package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"
)

const maxInputSize = 8 * 1024 // 8KB

func main() {
    var serverAddr string
    flag.StringVar(&serverAddr, "server", "http://localhost:8080", "server base URL (including scheme)")
    flag.Parse()

    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Enter username: ")
    user, _ := reader.ReadString('\n')
    user = strings.TrimSpace(user)
    if user == "" {
        fmt.Println("username required")
        os.Exit(1)
    }

    printHelp()

    for {
        fmt.Print("> ")
        line, err := reader.ReadString('\n')
        if err == io.EOF {
            fmt.Println("bye")
            return
        }
        if err != nil {
            log.Printf("read error: %v", err)
            return
        }
        line = strings.TrimRight(line, "\r\n")
        if len(line) > maxInputSize {
            fmt.Printf("input too long (max %d bytes), truncated\n", maxInputSize)
            line = line[:maxInputSize]
        }

        // commands
        lower := strings.ToLower(line)
        switch {
        case lower == "/exit" || lower == "exit":
            fmt.Println("exiting")
            return
        case lower == "/help" || lower == "help":
            printHelp()
            continue
        case strings.HasPrefix(lower, "/hist") || strings.HasPrefix(lower, "/history"):
            // parse optional sinceID
            parts := strings.Fields(line)
            var sinceID int64
            if len(parts) > 1 {
                id, err := strconv.ParseInt(parts[1], 10, 64)
                if err == nil {
                    sinceID = id
                } else {
                    fmt.Println("invalid sinceID, must be integer")
                    continue
                }
            }
            if err := fetchHistory(serverAddr, user, sinceID); err != nil {
                fmt.Printf("error fetching history: %v\n", err)
            }
            continue
        case lower == "/whoami":
            fmt.Println(user)
            continue
        }

        // otherwise send as message
        if err := sendMessage(serverAddr, user, line); err != nil {
            fmt.Printf("send error: %v\n", err)
        }
    }
}

func printHelp() {
    fmt.Println("Commands:")
    fmt.Println("  /help or help       - show this help")
    fmt.Println("  /history [sinceID]  - fetch your history (sinceID optional)")
    fmt.Println("  /whoami             - show current username")
    fmt.Println("  /exit or exit       - quit")
    fmt.Println("Type any other text to send it as a message.")
}

func sendMessage(server, author, text string) error {
    args := map[string]interface{}{"Author": author, "Text": text}
    b, _ := json.Marshal(args)
    resp, err := http.Post(server+"/send", "application/json", bytes.NewReader(b))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    var reply struct {
        OK    bool
        Error string
    }
    if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
        return err
    }
    if !reply.OK {
        return fmt.Errorf("server error: %s", reply.Error)
    }
    fmt.Println("sent")
    return nil
}

func fetchHistory(server, author string, sinceID int64) error {
    args := map[string]interface{}{"Author": author, "SinceID": sinceID}
    b, _ := json.Marshal(args)
    resp, err := http.Post(server+"/history", "application/json", bytes.NewReader(b))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    var reply struct {
        Messages []struct {
            ID        int64
            Author    string
            Text      string
            Timestamp time.Time
        }
    }
    if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
        return err
    }
    if len(reply.Messages) == 0 {
        fmt.Println("(no messages)")
        return nil
    }
    for _, m := range reply.Messages {
        ts := m.Timestamp.UTC().Format(time.RFC3339)
        fmt.Printf("%d %s %s\n", m.ID, ts, m.Text)
    }
    return nil
}
