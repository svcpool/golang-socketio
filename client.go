package gosocketio

import (
	"net"
	"strconv"
	"fmt"

	"github.com/ambelovsky/gosf-socketio/transport"
	"github.com/ambelovsky/gosf-socketio/protocol"
)

const (
	webSocketProtocol       = "ws://"
	webSocketSecureProtocol = "wss://"
	socketioUrl             = "/socket.io/?EIO=3&transport=websocket"
)

/**
Socket.io client representation
*/
type Client struct {
	methods
	Channel
	url string
}

type DialParams struct {
	Nsp string
}

/**
Get ws/wss url by host and port
*/
func GetUrl(host string, port int, secure bool) string {
	var prefix string
	if secure {
		prefix = webSocketSecureProtocol
	} else {
		prefix = webSocketProtocol
	}
	return prefix + net.JoinHostPort(host, strconv.Itoa(port)) + socketioUrl
}

/**
connect to host and initialise socket.io protocol

The correct ws protocol url example:
ws://myserver.com/socket.io/?EIO=3&transport=websocket

You can use GetUrlByHost for generating correct url
*/
func Dial(url string, tr transport.Transport, params DialParams) (*Client, error) {
	c := &Client{}
	c.url = url
	c.initChannel()
	c.initMethods()

	var err error
	c.conn, err = tr.Connect(url)
	if params.Nsp != "" {
		nspMsg := fmt.Sprintf("4%d%s", protocol.MessageTypeOpen, params.Nsp)
		c.conn.WriteMessage(nspMsg)
	}
	if err != nil {
		return nil, err
	}

	go inLoop(&c.Channel, &c.methods)
	go outLoop(&c.Channel, &c.methods)
	go pinger(&c.Channel)

	return c, nil
}

func Redial(c *Client) {
	var err error
	tr := transport.GetDefaultWebsocketTransport()
	for {
		c.conn, err = tr.Connect(c.url)
		if err == nil {
			break
		}
	}
	go inLoop(&c.Channel, &c.methods)
	go outLoop(&c.Channel, &c.methods)
	go pinger(&c.Channel)
}

/**
Close client connection
*/
func (c *Client) Close() {
	closeChannel(&c.Channel, &c.methods)
}
