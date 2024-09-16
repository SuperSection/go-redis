package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"time"

	"github.com/SuperSection/go-redis/client"
)

const defaultListenAddr = ":5001"

type Config struct {
	ListenAddr string
}

type Server struct {
	Config
	peers     map[*Peer]bool
	listener  net.Listener
	addPeerCh chan *Peer
	quitCh    chan struct{}
	msgCh     chan []byte

	keyval *KeyVal
}

/* Create new Server instance */
func NewServer(config Config) *Server {
	if len(config.ListenAddr) == 0 {
		config.ListenAddr = defaultListenAddr
	}

	return &Server{
		Config:    config,
		peers:     make(map[*Peer]bool),
		addPeerCh: make(chan *Peer),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan []byte),
		keyval:    NewKeyVal(),
	}
}

/* Start the Server */
func (server *Server) Start() error {
	ln, err := net.Listen("tcp", server.ListenAddr)
	if err != nil {
		return err
	}
	server.listener = ln

	go server.loop()

	slog.Info("server started", "listenAddress", server.ListenAddr)

	return server.acceptLoop()
}

/* handle incoming messages */
func (server *Server) handleRawMessage(rawMsg []byte) error {
	cmd, err := parseCommand(string(rawMsg))
	if err != nil {
		return err
	}

	switch v := cmd.(type) {
	case SetCommand:
		return server.keyval.Set(v.key, v.val)
	}

	return nil
}

/* Main Event Loop that runs in a separate goroutine */
func (server *Server) loop() {
	for {
		select {
		case rawMsg := <-server.msgCh:
			if err := server.handleRawMessage(rawMsg); err != nil {
				slog.Error("raw message error", "err", err)
			}
		case <-server.quitCh:
			return
		case peer := <-server.addPeerCh:
			server.peers[peer] = true
		}
	}
}

/* continuously Accepts New incoming connections */
func (server *Server) acceptLoop() error {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			slog.Error("accept error", "err", err)
			continue
		}
		go server.handleConn(conn)
	}
}

/* handling each new connection */
func (server *Server) handleConn(conn net.Conn) {
	peer := NewPeer(conn, server.msgCh)
	server.addPeerCh <- peer

	slog.Info("new peer connected", "remoteAddress", conn.RemoteAddr())

	go peer.readLoop()

	if err := peer.readLoop(); err != nil {
		slog.Error("peer read error", "err", err, "remoteAddress", conn.RemoteAddr())
	}
}

func main() {
	server := NewServer(Config{})

	go func() {
		log.Fatal(server.Start())
	}()

	time.Sleep(time.Second)

	client := client.NewClient("localhost:5001")

	for i := 0; i < 10; i++ {
		if err := client.Set(context.TODO(), fmt.Sprintf("foo_%d", i), fmt.Sprintf("bar_%d", i)); err != nil {
			log.Fatal(err)
		}
	}

	time.Sleep(time.Second)

	fmt.Println(server.keyval.data)
}
