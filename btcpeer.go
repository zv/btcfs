package main

import (
	"fmt"
	"github.com/conformal/btcwire"
	"log"
	"net"
)

const (
	MIN_PROTO_VERSION = 209

	CLIENT_VERSION = "btcfs"
)

type BTCPeer struct {
	Host, Port string
	Msgs chan btcwire.Message
	Quit chan bool

	Net btcwire.BitcoinNet
	Ver uint32

	con net.Conn
	conf Config
}

func NewBTCPeer(host, port string, conf Config) *BTCPeer {
	n := &BTCPeer{Host: host, Port: port, conf: conf}
	n.Msgs = make(chan btcwire.Message, 10)
	n.Quit = make(chan bool, 1)
	n.Net = btcwire.MainNet
	n.Ver = btcwire.ProtocolVersion
	return n
}

func (n *BTCPeer) Connect() error {
	peer := n.Host + ":" + n.Port
	conn, err := net.Dial("tcp", peer)
	if err != nil {
		return err
	}

	n.con = conn
	return nil
}

// Exchange version with remote peer.
// Error if they don't accept us
func (n *BTCPeer) DoVersion() error {
	nonce, err := btcwire.RandomUint64()
	if err != nil {
		return err
	}

	out, err := btcwire.NewMsgVersionFromConn(n.con, nonce, CLIENT_VERSION, n.conf.GetLastBlock())
	if err != nil {
		return err
	}

	if err := n.Write(out); err != nil {
		return err
	}

	// wait for version
version:
	for {
		in, err := n.Read()
		if err != nil {
			return err
		}

		switch in.(type) {
		case *btcwire.MsgVersion:
			msgver := in.(*btcwire.MsgVersion)
			if msgver.ProtocolVersion < MIN_PROTO_VERSION {
		return fmt.Errorf("peer protocol is %d, need >%d", msgver.ProtocolVersion, MIN_PROTO_VERSION)
			}
			// Send verack
			ack := btcwire.NewMsgVerAck()
			n.Write(ack)
			break version
		default:
			n.Close()
			return fmt.Errorf("expected version, got %s, aborting", in.Command())
		}
	}

verack:
	for {
		in, err := n.Read()
		if err != nil {
			return err
		}

		switch in.(type) {
		case *btcwire.MsgVerAck:
			// proceed
			break verack
		default:
			n.Close()
			return fmt.Errorf("expected verack, got %s, aborting", in.Command())
		}
	}

	return nil
}

// Close conn etc
func (n *BTCPeer) Close() {
	n.Quit <- true
	n.con.Close()
}

// Write one message
func (n *BTCPeer) Write(out btcwire.Message) error {
	log.Printf("%s:%s write %s", n.Host, n.Port, out.Command())
	return btcwire.WriteMessage(n.con, out, n.Ver, n.Net)
}

// Read one message
func (n *BTCPeer) Read() (btcwire.Message, error) {
	in, _, err := btcwire.ReadMessage(n.con, n.Ver, n.Net)
	if err != nil {
		return nil, err
	}

	log.Printf("%s:%s read %s", n.Host, n.Port, in.Command())

	return in, nil
}
