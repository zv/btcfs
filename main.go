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
				//	log.Print(err)
				}
			}

      log.Printf("%#v", blockchain) 
			
			locator := blockchain.CreateLocator()	
			log.Printf("locator: %#v", locator)
			getheaders := btcwire.NewMsgGetHeaders()

			for _, l := range locator {	 
				getheaders.AddBlockLocatorHash(&l)	
			}

			from.Write(getheaders)
			
			/*
			roots := make([]btcwire.ShaHash)	
		
			for _, h := range hdrs.Headers {
				buf := bytes.Buffer{}
				enc := gob.NewEncoder(&buf)
				enc.Encode(*h)
				val := buf.Bytes()
				//buf.Reset()
				//enc.Encode(h.Nonce)
				//key := buf.Bytes()
				//buf.Reset()
				db.Put(h.MerkleRoot.String(), val)
				getheaders.AddBlockLocatorHash(&h.MerkleRoot)
			}

			last_headers := hdrs.Headers[len(hdrs.Headers)-11:len(hdrs.Headers)-1]
			//last := hdrs.Headers[0]
			//getheaders := btcwire.NewMsgGetHeaders()
			fmt.Printf("%s\n", last_headers[0].MerkleRoot.String())
			for _, header := range last_headers{ 
				getheaders.AddBlockLocatorHash(&header.MerkleRoot)
				//fmt.Printf("hash %d: %s\n", i, header.MerkleRoot.String())
			}
			//fmt.Printf("%#v", getheaders)
			from.Write(getheaders)
			/*
			buf := bytes.Buffer{}
			enc := gob.NewEncoder(&buf)
			enc.Encode(&last.Nonce)
			nonce := buf.Bytes()	
			db.Put(string("Hello"), []byte("World"))
			test, _ := db.Get(string("Hello"))
			fmt.Printf("%s\n", test)
			fmt.Printf("%d Records in the DB\n", n)
			*/
			
		default:
	}
}
