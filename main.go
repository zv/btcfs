package main

import (
	"fmt"
	"github.com/conformal/btcwire"
	"code.google.com/p/leveldb-go/leveldb/memdb"
	"code.google.com/p/leveldb-go/leveldb/db"
	"log"
	"bytes"
	"encoding/gob"
)

var (
	conf = Config{"btcfs", 0}

	headerdb db.DB
)

func init() {
	gob.Register(btcwire.BlockHeader{})
	headerdb = memdb.New(nil)
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
			hdrs := data.(*btcwire.MsgHeaders)

			log.Printf("Received %d headers", len(hdrs.Headers))

			for _, h := range hdrs.Headers {
				buf := bytes.Buffer{}
				enc := gob.NewEncoder(&buf)
				enc.Encode(*h)
				val := buf.Bytes()
				buf.Reset()
				enc.Encode(h.MerkleRoot)
				key := buf.Bytes()
				buf.Reset()
				headerdb.Set(key, val, nil)
			}

			last := hdrs.Headers[len(hdrs.Headers)-1]
			getheaders := btcwire.NewMsgGetHeaders()
			getheaders.AddBlockLocatorHash(&last.MerkleRoot)
			from.Write(getheaders)
			
		default:
	}
}