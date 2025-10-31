package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "go_chat_bot/pkg/chat"
)

func main() {
    var port int
    var logfile string
    flag.IntVar(&port, "port", 8080, "port to listen on")
    flag.StringVar(&logfile, "logfile", "", "optional logfile path")
    flag.Parse()

    if logfile != "" {
        f, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
        if err != nil {
            log.Fatalf("failed to open logfile: %v", err)
        }
        log.SetOutput(f)
        defer f.Close()
    }

    svc := chat.NewChatService()

    mux := http.NewServeMux()

    mux.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        var args chat.SendMessageArgs
        dec := json.NewDecoder(r.Body)
        if err := dec.Decode(&args); err != nil {
            log.Printf("ERROR method=send reason=bad-request remote=%s err=%v", r.RemoteAddr, err)
            http.Error(w, "bad request", http.StatusBadRequest)
            return
        }

        // Log connect-like info
        log.Printf("REQUEST method=send client=%s remote=%s", args.Author, r.RemoteAddr)

        var reply chat.SendMessageReply
        if err := svc.SendMessage(args, &reply); err != nil {
            w.WriteHeader(http.StatusBadRequest)
        }
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        _ = json.NewEncoder(w).Encode(reply)
    })

    mux.HandleFunc("/history", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        var args chat.GetHistoryArgs
        if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
            log.Printf("ERROR method=history reason=bad-request remote=%s err=%v", r.RemoteAddr, err)
            http.Error(w, "bad request", http.StatusBadRequest)
            return
        }
        log.Printf("REQUEST method=history client=%s remote=%s since=%d", args.Author, r.RemoteAddr, args.SinceID)

        var reply chat.GetHistoryReply
        if err := svc.GetHistory(args, &reply); err != nil {
            w.WriteHeader(http.StatusBadRequest)
        }
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        _ = json.NewEncoder(w).Encode(reply)
    })

    srv := &http.Server{
        Addr:         fmt.Sprintf(":%d", port),
        Handler:      mux,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    // Start server
    go func() {
        log.Printf("START server port=%d", port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("server error: %v", err)
        }
    }()

    // Graceful shutdown on interrupt
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
    <-stop

    log.Printf("SHUTDOWN server - waiting up to 5s for connections to close")
    _ = srv.Close()
    time.Sleep(250 * time.Millisecond)
    log.Printf("SHUTDOWN complete")
}
