package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"

	"go_chat_bot/pkg/chat"
)

const maxInputSize = 8 * 1024 // 8KB

func main() {
	var serverAddr string
	flag.StringVar(&serverAddr, "server", "localhost:8080", "server address and port")
	flag.Parse()

	// Connect to RPC server
	client, err := rpc.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter username: ")
	user, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("error reading username: %v", err)
	}
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
			if err := fetchHistory(client, user, sinceID); err != nil {
				fmt.Printf("error fetching history: %v\n", err)
			}
			continue
		case lower == "/whoami":
			fmt.Println(user)
			continue
		}

		// otherwise send as message
		if err := sendMessage(client, user, line); err != nil {
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

func sendMessage(client *rpc.Client, author, text string) error {
	args := &chat.SendMessageArgs{
		Author: author,
		Text:   text,
	}
	var reply chat.SendMessageReply
	if err := client.Call("ChatService.SendMessage", args, &reply); err != nil {
		return err
	}
	if !reply.OK {
		return fmt.Errorf("server error: %s", reply.Error)
	}
	fmt.Println("sent")
	return nil
}

func fetchHistory(client *rpc.Client, author string, sinceID int64) error {
	args := &chat.GetHistoryArgs{
		Author:  author,
		SinceID: sinceID,
	}
	var reply chat.GetHistoryReply
	if err := client.Call("ChatService.GetHistory", args, &reply); err != nil {
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
