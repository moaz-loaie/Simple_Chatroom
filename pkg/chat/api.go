package chat

import (
	"errors"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

// Message represents a single chat message.
type Message struct {
	ID        int64     // unique message id
	Author    string    // username or client id
	Text      string    // message body
	Timestamp time.Time // when the server recorded the message
}

// SendMessageArgs are provided by clients when sending a message.
type SendMessageArgs struct {
	Author string
	Text   string
}

// SendMessageReply is returned after sending a message.
type SendMessageReply struct {
	OK    bool
	Error string
}

// GetHistoryArgs requests messages for a specific Author.
type GetHistoryArgs struct {
	Author  string
	SinceID int64 // optional, 0 means all
}

// GetHistoryReply contains the messages returned to the client.
type GetHistoryReply struct {
	Messages []Message
}

// ChatService implements the RPC methods for the chatroom.
// It is safe for concurrent use.
type ChatService struct {
	mu       sync.RWMutex
	nextID   int64
	storage  []Message
	listener net.Listener // for graceful shutdown
}

// NewChatService creates and returns a ready-to-use ChatService.
func NewChatService() *ChatService {
	return &ChatService{
		nextID:  1,
		storage: make([]Message, 0, 64),
	}
}

// RegisterService registers the ChatService with an RPC server.
func (s *ChatService) RegisterService() error {
	return rpc.RegisterName("ChatService", s)
}

// ServeRPC starts listening for RPC connections on the given network and address.
// It can be stopped by calling Shutdown. The function will return with nil error
// when shutdown is called, or with an error if there's a failure in accepting
// new connections.
func (s *ChatService) ServeRPC(network, addr string) error {
	var err error
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listener, err = net.Listen(network, addr)
	if err != nil {
		return err
	}
	// Release the lock after we've created the listener
	s.mu.Unlock()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// Check if error is due to closed listener
			if err, ok := err.(*net.OpError); ok && err.Op == "accept" {
				return nil // Return without error as this is a normal shutdown
			}
			log.Printf("accept error: %v", err)
			return err
		}
		go rpc.ServeConn(conn)
	}
}

// SendMessage appends a message authored by args.Author to the server history.
func (s *ChatService) SendMessage(args SendMessageArgs, reply *SendMessageReply) error {
	if reply == nil {
		return errors.New("internal: nil reply")
	}
	if args.Author == "" {
		reply.OK = false
		reply.Error = "author required"
		log.Printf("ERROR method=SendMessage reason=missing-author")
		return errors.New(reply.Error)
	}
	if args.Text == "" {
		// Allow empty text but log
		log.Printf("SEND client=%s text=<empty>", args.Author)
	}

	s.mu.Lock()
	id := s.nextID
	s.nextID++
	msg := Message{
		ID:        id,
		Author:    args.Author,
		Text:      args.Text,
		Timestamp: time.Now().UTC(),
	}

	// Keep only the last 1000 messages to prevent unbounded growth
	if len(s.storage) >= 1000 {
		s.storage = append(s.storage[1:], msg)
	} else {
		s.storage = append(s.storage, msg)
	}
	s.mu.Unlock()

	log.Printf("SEND client=%s id=%d text=%s", args.Author, msg.ID, args.Text)

	reply.OK = true
	reply.Error = ""
	return nil
}

// Shutdown gracefully shuts down the RPC server by closing the listener.
func (s *ChatService) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
	}
}

// GetHistory returns messages authored by args.Author. If SinceID>0 only messages
// with ID > SinceID are returned.
func (s *ChatService) GetHistory(args GetHistoryArgs, reply *GetHistoryReply) error {
	if reply == nil {
		return errors.New("internal: nil reply")
	}
	if args.Author == "" {
		log.Printf("ERROR method=GetHistory reason=missing-author")
		return errors.New("author required")
	}

	s.mu.RLock()
	// Make a copy of matching messages to avoid holding lock while returning.
	var out []Message
	for _, m := range s.storage {
		if m.Author != args.Author {
			continue
		}
		if args.SinceID > 0 && m.ID <= args.SinceID {
			continue
		}
		out = append(out, m)
	}
	s.mu.RUnlock()

	// Return results
	reply.Messages = out
	log.Printf("HISTORY client=%s returned=%d since=%d", args.Author, len(out), args.SinceID)
	return nil
}
