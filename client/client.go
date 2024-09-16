package client

import (
	"bytes"
	"context"
	"io"
	"log"
	"net"

	"github.com/tidwall/resp"
)

type Client struct {
	addr string
	conn net.Conn
}

func NewClient(addr string) *Client {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	return &Client{
		addr: addr,
		conn: conn,
	}
}

func (client *Client) Set(ctx context.Context, key string, val string) error {

	buf := &bytes.Buffer{}
	wr := resp.NewWriter(buf)

	wr.WriteArray([]resp.Value{
		resp.StringValue("SET"),
		resp.StringValue(key),
		resp.StringValue(val),
	})

	_, err := io.Copy(client.conn, buf)

	return err
}

func (client *Client) Get(ctx context.Context, key string) (string, error) {

	buf := &bytes.Buffer{}
	wr := resp.NewWriter(buf)

	wr.WriteArray([]resp.Value{
		resp.StringValue("GET"),
		resp.StringValue(key),
	})

	_, err := io.Copy(client.conn, buf)
	if err != nil {
		return "", err
	}

	b := make([]byte, 1024)
	n, err := client.conn.Read(b)

	return string(b[:n]), err
}
