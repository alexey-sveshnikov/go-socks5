package socks5

import (
	"strings"
	"bytes"
	"time"
	"log"
	"context"
	"net"
	"testing"
	"strconv"
)

// Creates the server that only prints everything it receives
// Returns port number
func createTargetServer() (int) {
	l, err := net.Listen("tcp", "127.0.0.1:0");
	if err != nil {
		panic(err)
	}

	go func() {
		conn, err := l.Accept();
		if err != nil {
			panic(err)
		}
		//fmt.Printf("Test server accepted connection from %v\n", conn.RemoteAddr())

		for {
			buf := make([]byte, 32)
			_, err := conn.Read(buf);
			if err != nil {
				return
			}
			//fmt.Printf("Server got data: %s\n", string(buf))
		}
	}()

	_, targetPortStr, err := net.SplitHostPort(l.Addr().String());
	if err != nil {
		panic(err)
	}

	port, err := strconv.Atoi(targetPortStr);
	if err != nil {
		panic(err)
	}
	//fmt.Printf("Test server listens on port %d\n", port)
	return port
}

// Returns server connection with fake client
func getClientConnection() (net.Conn) {
	l, err := net.Listen("tcp", "127.0.0.1:0");
	if err != nil {
		panic(err)
	}

	go func() {
		conn, err := net.Dial("tcp", l.Addr().String());
		if err != nil {
			panic(err)
		}
		//fmt.Printf("Test client dialed to %s\n", l.Addr())
		conn.Write([]byte("Hello, world!\n"))
		for {
			buf := make([]byte, 32)
			_, err := conn.Read(buf);
			if err != nil {
				return
			}
			//fmt.Printf("Client got data: %v\n", string(buf))
		}
	}()

	conn, err := l.Accept();
	if err != nil {
		panic(err)
	}

	return conn
}

func TestStatEvents(t *testing.T) {
	clientConn := getClientConnection()
	destServerPort := createTargetServer()
	destAddr := &AddrSpec{
		IP:   net.ParseIP("127.0.0.1"),
		Port: destServerPort,
	}

	ctx := context.Background()
	req := &Request{
		Command: ConnectCommand,
		AuthContext: &AuthContext{
			Method: UserPassAuth,
			Payload: map[string]string{
				"Username": "TestUser",
			},
		},
		RemoteAddr: &AddrSpec{
			IP:   net.ParseIP("192.168.1.1"),
			Port: 10010,
		},
		DestAddr:     destAddr,
		realDestAddr: destAddr,
		bufConn:      clientConn,
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		clientConn.Close()
	}()

	logBuffer := &bytes.Buffer{}

	logger := log.New(logBuffer, "", log.LstdFlags)
	logger.Println("test")
	s := &Server{
		config: &Config{
			Rules:    PermitAll(),
			Resolver: DNSResolver{},
			Logger:   logger,
		},
		eventDispatcher: EventDispatcher{[]EventsHandler{
			LoggingEventsHandler{logger},
		}},
	}
	err := s.handleConnect(ctx, clientConn, req)
	if !strings.Contains(err.Error(), "closed network connection") {
		t.Fatal(err)
	}
	//fmt.Println("Log contents:")
	//fmt.Println(logBuffer)

	if !strings.Contains(logBuffer.String(), "TestUser connected") {
		t.Fatal("No connected message")
	}

	if !strings.Contains(logBuffer.String(), "TestUser connect 127.0.0.1:") {
		t.Fatal("No connect message")
	}

	if !strings.Contains(logBuffer.String(), "TestUser uploaded 14 bytes") {
		t.Fatal("No uploaded message")
	}

	if !strings.Contains(logBuffer.String(), "TestUser disconnected, session length: 0.1") {
		t.Fatal("No disconnect message")
	}

}
