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

	for {
		in, err := peer.Read()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%#v", in)
	}

}

