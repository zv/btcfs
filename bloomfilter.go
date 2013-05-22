package main

import (
	"fmt"
	"github.com/conformal/btcwire"
	"log"
)

const (
	BitcoinMurmurMagic = 0xFBA4C795
)

// Bitcoin uses version 3 of the 32-bit Murmur hash function.
// To get N "different" hash functions we simply initialize the Murmur algorithm with the following formula: (HashFuncsN * 0xFBA4C795 (BitcoinMurmurMagic) + SeedTweak)
// if the filter is initialized with 4 hash functions and a tweak of 0x00000005, when the second function (index 1) is needed h1 would be equal to 4221880218.
func (msg *MsgGetBlocks) GenerateHashFunctions() {
	n := &msg.HashFuncsN
	s := &msg.SeedTweak
	n*BitcoinMurmurMagic + s
}

type MsgFilterLoad struct {
	Filter      [1024]uint8
	HashFuncsN  uint32 // number of hash functions used in this Bloom filter
	SeedTweak   uint32 // A random value to add to the seed value in the hash function used by the bloom filter.
	FlagControl uint8  // A set of flags that control how matched items are added to the filte
}

// When loading a filter with the filterload command, there are two parameters that can be chosen.
// One is the size of the filter in bytes (FilterSize).
// The other is the number of hash functions to use (DeriveHashFuncN).


//The number of hash functions required is given by S * 8 / N * log(2).
func (msg *MsgGetBlocks) DeriveHashFuncN(s int) {
	n := &msg.HashFuncsN
	return (s * 8) / (n * log(2))
}

// P represents the probability of a false positive, where 1.0 is "match everything" and zero is unachievable.
func (msg *MsgGetBlocks) DeriveFilterSize(p int) {
	n := &msg.HashFuncsN

	//The size S of the filter in bytes is given by (-1 / pow(log(2), 2) * N * log(P)) / 8
	s := (-1/ (log(2) ^ 2) * n * log(p)) / 8

	// 36,000: selected as it represents a filter of 20,000 items with false positive
	// rate of < 0.1% or 10,000 items and a false positive rate of < 0.0001%
	if s > 32000 {
		return 32000
	}

	return s
}


// Implementation of wikipedia Murmur 3 32 bit hashing
// http://en.wikipedia.org/w/index.php?title=MurmurHash&oldid=551912607#Algorithm
func Murmur3_32(key []byte, seed uint32) {
	var (
		c1 uint32 = 0xcc9e2d51
		c2 uint32 = 0x1b873593
		r1 uint32 = 15
		r2 uint32 = 13
		m  uint32 = 5
		n  uint32 = 0xe6546b64
    length uint32 = len(key)

		k uint32
	)

	hash := seed
	nblocks := length / 4
	buf := bytes.NewBuffer(key)
	for _ := range nblocks {
		binary.Read(buf, binary.LittleEndian, &k)
		k *= c1
		k = (k << r1) | (k >> (32 - r1))
		k *= c2
		hash ^= k
		hash = (hash << r2) | (hash >> (32 - r2))
		hash = (hash * m) + n
	}

	k = 0
	tailIndex := nblocks * 4
	switch length & 3 {
	case 3:
		k ^= uint32(key[tailIndex+2]) << 16
		fallthrough
	case 2:
		k ^= uint32(key[tailIndex+1]) << 8
		fallthrough
	case 1:
		k ^= uint32(key[tailIndex])
		k *= c1
		k = (k << r2) | (k >> (32 - r1))
		k *= c2
		hash ^= k
	}
	hash ^= uint32(length)
	hash ^= hash >> 16
	hash *= 0x85ebca6b
	hash ^= hash >> r2
	hash *= 0xc2b2ae35
	hash ^= hash >> 16

	return hash
}
