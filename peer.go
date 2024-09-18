package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/tidwall/resp"
)

/* Peer represents a single client connection (a "peer") to the server */
type Peer struct {
	conn  net.Conn
	msgCh chan Message
}

func (p *Peer) Send(msg []byte) (int, error) {
	return p.conn.Write(msg)
}

/* constructor function to initialize a new Peer instance */
func NewPeer(conn net.Conn, msgCh chan Message) *Peer {
	return &Peer{
		conn:  conn,
		msgCh: msgCh,
	}
}

/* continuously reads data from the peer's connection */
func (p *Peer) readLoop() error {
	rd := resp.NewReader(p.conn)

	for {
		v, _, err := rd.ReadValue()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if v.Type() == resp.Array {
			for _, value := range v.Array() {
				switch value.String() {
				case CommandSET:
					if len(v.Array()) != 3 {
						return fmt.Errorf("invalid number of variables for SET command")
					}

					cmd := SetCommand{
						key: v.Array()[1].Bytes(),
						val: v.Array()[2].Bytes(),
					}
					// fmt.Printf("got SET cmd %+v\n", cmd)

					p.msgCh <- Message{
						cmd:  cmd,
						peer: p,
					}

				case CommandGET:
					if len(v.Array()) != 2 {
						return fmt.Errorf("invalid number of variables for GET command")
					}

					cmd := GetCommand{
						key: v.Array()[1].Bytes(),
					}
					// fmt.Printf("got GET cmd %+v\n", cmd)

					p.msgCh <- Message{
						cmd:  cmd,
						peer: p,
					}
				}
			}
		}
	}

	return nil
}
