package main

import (
	//	"code.google.com/p/gocask"
	"encoding/gob"
	"fmt"
	"github.com/conformal/btcwire"
	"log"
)

var (
	conf = Config{"btcfs", 0}
	//	db, _      = gocask.NewGocask(".")
	getheaders      = btcwire.NewMsgGetHeaders()
	blockchain      = InitializeBlockChain()
	blockheaderchan = make(chan *btcwire.BlockHeader)
)

func init() {
	gob.Register(btcwire.BlockHeader{})
	getheaders.AddBlockLocatorHash(&btcwire.GenesisMerkleRoot)
}

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

	peer.Write(getheaders)

	go func() {
		err := SrvHeaders(blockheaderchan)
		if err != nil {
			log.Printf("SrvHeaders failed: %s", err)
		}
	}()

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
	//	defer db.Close()

	switch msg {
	case "headers":
		hdrs := data.(*btcwire.MsgHeaders)
		log.Printf("Received %d headers", len(hdrs.Headers))

		for _, h := range hdrs.Headers {
			blockheaderchan <- h
			_, err := blockchain.AddBlock(h)
			if err != nil {
				log.Print(err)
			}
		}

		//blockchain.Genesis.PrintSubtree(1)

		log.Printf("Blockchain Head Depth: %d\n", blockchain.ChainHeadDepth)
		log.Printf("Chain Head: %#v\n", blockchain.ChainHead.Hash.String())

		getheaders := btcwire.NewMsgGetHeaders()

		locator := blockchain.CreateLocator()
		getheaders.BlockLocatorHashes = locator

		from.Write(getheaders)

	default:
	}
}
