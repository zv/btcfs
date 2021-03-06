package main

import (
	//	"code.google.com/p/gocask"
	"encoding/gob"
	"fmt"
	"github.com/conformal/btcwire"
	"log"
	"time"
  "os"
  "os/signal"
)

var (
	conf = Config{"btcfs", 0}
	//	db, _           = gocask.NewGocask(".")
	getheaders      = btcwire.NewMsgGetHeaders()
	blockchain, dberr = InitializeBlockChain()
	blockheaderchan = make(chan *btcwire.BlockHeader)
)

func init() {
	gob.Register(btcwire.BlockHeader{})
	getheaders.AddBlockLocatorHash(&btcwire.GenesisMerkleRoot)

  c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt)
  go func(){
      for sig := range c {
        blockchain.Database.Close()
        fmt.Println("Closed Database\n")
        fmt.Println(sig)
        return
      }
  }()
}

func main() {
	addrs, err := FindPeers(10)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("haters gonna hate")
	vchan := make(chan bool)

	for _, addr := range addrs {
		peer := NewBTCPeer(addr, btcwire.MainPort, conf)
		go func() {
			if err := peer.Connect(); err != nil {
				log.Fatal(err)
			}
			if err := peer.DoVersion(); err != nil {
				log.Fatal(err)
			}
			fmt.Println("haters gonna hate")

			peer.Write(getheaders)
			fmt.Println("haters gonna hate")

			go func() {
				err := SrvHeaders(blockheaderchan)
				if err != nil {
					log.Printf("SrvHeaders failed: %s", err)
				}
			}()
			vchan <- true

		}()

		select {
		case <-vchan:
		case <-time.After(5 * time.Second):
			continue
		}

		ProcessMessages(peer)
	}
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
		hdrs := data.(*btcwire.MsgHeaders)
		log.Printf("Received %d headers", len(hdrs.Headers))

		for _, h := range hdrs.Headers {
			blockheaderchan <- h
			_, err := blockchain.InitializeBlock(h)
			if err != nil {
				log.Print(err)
			}
		}

		// blockchain.Genesis.PrintSubtree(1)

		log.Printf("Blockchain Head Depth: %d\n", blockchain.ChainHeadDepth)
		log.Printf("Chain Head: %#v\n", blockchain.ChainHead.String())

		getheaders := btcwire.NewMsgGetHeaders()

		locator := blockchain.CreateLocator()
		getheaders.BlockLocatorHashes = locator

		from.Write(getheaders)

	default:
	}
}
