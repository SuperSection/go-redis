package client

import (
	"bytes"
	"context"
	"net"

	"github.com/tidwall/resp"
)

type Client struct {
	addr string
	conn net.Conn
}

func NewClient(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Client{
		addr: addr,
		conn: conn,
	}, nil
}

func (client *Client) Set(ctx context.Context, key string, val string) error {

	var buf bytes.Buffer
	wr := resp.NewWriter(&buf)

	wr.WriteArray([]resp.Value{
		resp.StringValue("SET"),
		resp.StringValue(key),
		resp.StringValue(val),
	})

	_, err := client.conn.Write(buf.Bytes())

	return err
}

func (client *Client) Get(ctx context.Context, key string) (string, error) {

	var buf bytes.Buffer
	wr := resp.NewWriter(&buf)

	wr.WriteArray([]resp.Value{
		resp.StringValue("GET"),
		resp.StringValue(key),
	})

	_, err := client.conn.Write(buf.Bytes())
	if err != nil {
		return "", err
	}

	b := make([]byte, 1024)
	n, err := client.conn.Read(b)

	return string(b[:n]), err
}
