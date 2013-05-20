package main

import (
	"fmt"
	"github.com/conformal/btcwire"
	"code.google.com/p/gocask"
	"log"
	//"bytes"
	"encoding/gob"
)

var (
	conf = Config{"btcfs", 0}
	db, _ = gocask.NewGocask(".")
	getheaders = btcwire.NewMsgGetHeaders()
	blockchain = InitializeBlockChain()  
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

	//headerschan := make(chan btcwire.BlockHeader, 20)
	

	peer.Write(getheaders)

	ProcessMessages(peer)

	/*
	iter := reader.Find(nil, nil)
	"code.google.com/p/leveldb-go/leveldb/table"
	for iter.Next() {
		fmt.Printf("nonce: %q, val: %q\n", iter.Key(), iter.Value())
	} 
	iter.Close()
	*/
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
	defer db.Close()

	switch msg {
		case "headers":
			hdrs := data.(*btcwire.MsgHeaders)
			log.Printf("Received %d headers", len(hdrs.Headers))

			for _, h := range hdrs.Headers {
				_, err := blockchain.AddBlock(h)
				if err != nil {
					log.Print(err)
				}
			}

      log.Printf("Blockchain Head Depth: %d\n", blockchain.ChainHeadDepth) 
			
			locator := blockchain.CreateLocator()	
			log.Printf("locator: %#v", locator)
			getheaders := btcwire.NewMsgGetHeaders()

      getheaders.BlockLocatorHashes = locator

			from.Write(getheaders)
			
		default:
	}
}
