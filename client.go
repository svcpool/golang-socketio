package gosocketio

import (
	"fmt"
	"net"
	"strconv"

	"github.com/ambelovsky/gosf-socketio/protocol"
	"github.com/ambelovsky/gosf-socketio/transport"
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
	url    string
	params DialParams
	tr     transport.Transport
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
	c.params = params
	c.tr = tr
	c.initChannel()
	c.initMethods()
	c.Client = c

	var err error
	c.conn, err = tr.Connect(url)
	if err != nil {
		return nil, err
	}
	if params.Nsp != "" {
		nspMsg := fmt.Sprintf("4%d%s", protocol.MessageTypeOpen, params.Nsp)
		c.conn.WriteMessage(nspMsg)
		c.Emit(params.Nsp, nil)
	}

	go inLoop(&c.Channel, &c.methods)
	go outLoop(&c.Channel, &c.methods)
	go pinger(&c.Channel)

	return c, nil
}

func Redial(c *Client) (*Client, error) {
	var err error
	c, err = Dial(c.url, c.tr, c.params)

	return c, err
}

/**
Close client connection
*/
func (c *Client) Close() {
	closeChannel(&c.Channel, &c.methods)
}
