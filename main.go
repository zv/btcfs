package main

import (
	"fmt"
	"github.com/conformal/btcwire"
	"log"
)

var (
	conf = Config{"btcfs", 0}
)

func main() {
	addrs, err := FindPeers(1)
	if err != nil {
		fmt.Println(err)
		return
	}

	peer := NewBTCPeer(addrs[0], btcwire.MainPort, conf)

	if err := peer.Connect(); err != nil {
		log.Fatal(err)
	}

	if err := peer.DoVersion(); err != nil {
		log.Fatal(err)
	}

	//headerschan := make(chan btcwire.BlockHeader, 20)

	getheaders := btcwire.NewMsgGetHeaders()
	getheaders.AddBlockLocatorHash(&btcwire.GenesisMerkleRoot)

	peer.Write(getheaders)

	ProcessMessages(peer)

}

func ProcessMessages(n *BTCPeer) error {
	for {
		in, err := n.Read()
		if err != nil {
			return fmt.Errorf("ProcessMessages: peer.Read: %s", err)
		}
		ProcessMessage(n, in.Command(), in)
	}

	return nil
}

func ProcessMessage(from *BTCPeer, msg string, data btcwire.Message) {
	//log.Printf("ProcessMessage: %s %#v", msg, data)

	switch msg {
		case "headers":
		default:
	}
}