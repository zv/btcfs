package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
)

var (
	dns []string = []string{
		"dnsseed.bluematt.me",
		"seed.bitcoin.sipa.be",
		"dnsseed.bitcoin.dashjr.org",
		"bitseed.xf2.org",
	}
)

// Find up to n peers from a dns server.
// return nil, err on error
func FindPeers(n int) ([]string, error) {
	return []string{"85.226.21.108"}, nil
	for _, rnd := range rand.Perm(len(dns)) {
		addrs, err := net.LookupHost(dns[rnd])
		if err != nil {
			log.Print(err)
			continue
		}

		if len(addrs) >= n {
			return addrs[:5], nil
		} else {
			return addrs, nil
		}

	}

	return nil, fmt.Errorf("No peers found.")
}
