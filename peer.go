package main

import (
	"log/slog"
	"net"
)

/* Peer represents a single client connection (a "peer") to the server */
type Peer struct {
	conn  net.Conn
	msgCh chan []byte
}

/* constructor function to initialize a new Peer instance */
func NewPeer(conn net.Conn, msgCh chan []byte) *Peer {
	return &Peer{
		conn:  conn,
		msgCh: msgCh,
	}
}

/* continuously reads data from the peer's connection */
func (p *Peer) readLoop() error {
	buf := make([]byte, 1024)

	for {
		n, err := p.conn.Read(buf)
		if err != nil {
			slog.Error("peer read error", "err", err)
			return err
		}
		// fmt.Println(string(buf[:n]))
		// fmt.Println(len(buf[:n]))
		msgBuf := make([]byte, n)
		copy(msgBuf, buf[:n])
		p.msgCh <- msgBuf
	}
}
