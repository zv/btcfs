package main

import (
	"bytes"
	"encoding/binary"
	"math"

//	"github.com/conformal/btcwire"
)

var (
	BitcoinMurmurMagic         uint32 = 0xFBA4C795
	MaxHashFunctions           uint32 = 50
	MaxFilterSize              uint32 = 32000
	bloom_mask                        = []uint8{0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80}
	BLOOM_UPDATE_NONE                 = 0
	BLOOM_UPDATE_ALL                  = 1
	BLOOM_UPDATE_P2PUBKEY_ONLY        = 2
	BLOOM_UPDATE_MASK                 = 3
)

type BloomFilter struct {
	Vector     []uint8 // bitvector
	nHashFuncs uint32  // number of hash functions used in this Bloom filter
	nTweak     uint32  // A random value to add to the seed value in the hash function used by the bloom filter.
	nFlag      uint8   // A set of flags that control how matched items are added to the filte
}

// When loading a filter with the filterload command, there are two parameters that can be chosen.
// One is the size of the filter in bytes (FilterSize).
// The other is the number of hash functions to use, taking the probability distribution of the bitvector (DeriveHashFuncN).

//The number of hash functions required is given by S * 8 / N * math.Log(2).
func (filter *BloomFilter) DeriveHashFunctionsN(size float64) int {
	n := float64(filter.nHashFuncs)
	hash_functions := (size * 8) / (n * math.Log(2))
	return int(math.Min(hash_functions, float64(MaxHashFunctions)))
}

// IdealFilterSize gives the ideal bloom filter size given a probability distribution p between 1.0 "match everything" and 0  (
// P represents the probability of a false positive, where 1.0 is "match everything" and zero is unachievable.
func (filter *BloomFilter) IdealFilterSize(p float64) int {
	n := float64(filter.nHashFuncs)

	//The ideal size S of the filter in bytes is given by (-1 / pow(math.Log(2), 2) * N * log(P)) / 8
	// The MaxFilterSize a filter of 20,000 items with false positive
	// rate of < 0.1% or 10,000 items and a false positive rate of < 0.0001%
	// which is more than enough
	size := -1 / (math.Exp2(math.Ln2) * n * math.Log(p))
	s := int(math.Min(size/8, float64(MaxFilterSize)))
	return s
}

// Implementation of wikipedia Murmur versoin 3 32 bit hashing,
// http://en.wikipedia.org/w/index.php?title=MurmurHash&oldid=551912607#Algorithm
func (filter *BloomFilter) BitcoinMurmur(key []byte, seed_int uint32) uint32 {

	var (
		c1     uint32 = 0xcc9e2d51
		c2     uint32 = 0x1b873593
		r1     uint32 = 15
		r2     uint32 = 13
		m      uint32 = 5
		n      uint32 = 0xe6546b64
		length uint32 = uint32(len(key))
		k      uint32
	)

	// To get N "different" hash functions we simply initialize the Murmur algorithm
	// with the following formula: (nHashFuncs[i] * 0xFBA4C795 (BitcoinMurmurMagic) + nTweak[i])
	tweak := filter.nTweak
	seed := seed_int*BitcoinMurmurMagic + tweak

	hash := seed
	nblocks := length / 4
	buf := bytes.NewBuffer(key)
	for i := uint32(0); i < nblocks; i++ {
		binary.Read(buf, binary.LittleEndian, &k)
		k *= c1
		k = (k << r1) | (k >> (32 - r1))
		k *= c2
		hash ^= k
		hash = (hash << r2) | (hash >> (32 - r2))
		hash = (hash * m) + n
	}

	k = 0
	switch length & 3 {
	case 3:
		k ^= uint32(key[length+2]) << 16
		fallthrough
	case 2:
		k ^= uint32(key[length+1]) << 8
		fallthrough
	case 1:
		k ^= uint32(key[length])
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

// Insert a new block key
func (filter *BloomFilter) Insert(key []byte) {

	if len(filter.Vector) == 1 && filter.Vector[0] == 0xff {
		return
	}

	for i := uint32(0); i < filter.nHashFuncs; i++ {
		indices := filter.BitcoinMurmur(key, i)

		// Sets the bit index of our Bloom vector
		filter.Vector[indices>>3] |= bloom_mask[7&indices]
	}
	return
}

// Check if our bloom filter contains a key
func (filter *BloomFilter) Contains(key []byte) bool {
	if len(filter.Vector) == 1 && key[0] == 0xff {
		return true
	}
	for i := uint32(0); i < filter.nHashFuncs; i++ {
		indices := filter.BitcoinMurmur(key, i)

		// checks the bit index of our Bloom vector
		if (filter.Vector[indices>>3] & bloom_mask[7&indices]) != 0 {
			return false
		}
	}
	return true
}
