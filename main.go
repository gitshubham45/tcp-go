package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type Message struct {
	from    string
	payload string
}

type Server struct {
	listenAddr string
	ln         net.Listener
	quitch     chan struct{}
	msgch      chan Message
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan struct{}),
		msgch:      make(chan Message, 10),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.ln = ln

	// Start the accept loop in a goroutine
	go s.acceptLoop()

	// Wait for a signal to quit
	<-s.quitch
	close(s.msgch)

	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println("error accepting:", err)
			return
		}

		fmt.Println("New connection to the server:", conn.RemoteAddr())

		// Start a goroutine to handle each connection
		go s.readLoop(conn)
	}
}

// func (s *Server) readLoop(conn net.Conn) {
// 	defer conn.Close()

// 	buf := make([]byte, 2048)

// 	for {
// 		// Read data until newline
// 		n, err := conn.Read(buf)
// 		if err != nil {
// 			fmt.Println("Read error:", err)
// 			break
// 		}

// 		s.msgch <- Message{
// 			from: conn.RemoteAddr().String(),
// 			payload: buf[:n],
// 		}
// 	}
// }

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		// Read data until newline
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Read error:", err)
			break
		}

		// Process the complete message
		s.msgch <- Message{
			from:    conn.RemoteAddr().String(),
			payload: message,
		}

		conn.Write([]byte("thank you for your message!"))
	}
}

func main() {
	server := NewServer(":3000")

	go func() {
		for msg := range server.msgch {
			fmt.Printf("Recieved message from connection  (%s):%s ", msg.from, msg.payload)
		}
	}()

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}

	// Simulate server running (e.g., wait for user input to stop)
	fmt.Println("Press Enter to stop the server...")
	fmt.Scanln()

	// Signal the server to shut down
	close(server.quitch)
}
