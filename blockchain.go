package main

import (
	"bytes"
	"code.google.com/p/gocask"
	"encoding/gob"
	"fmt"
	"github.com/conformal/btcwire"
	"log"
	"strings"
)

type BlockChain struct {
	Last           btcwire.ShaHash
	Genesis        btcwire.ShaHash
	ChainHead      btcwire.ShaHash
	ChainHeadDepth int
	Database       *gocask.Gocask
}

type BlockHandle struct {
	Hash   btcwire.ShaHash
	Header btcwire.BlockHeader
	Depth  int
	// We can't know the size of a block apriori
	// so lets store a key to lookup the block when we
	// need it 
	Block string
}

// return the block headers identifier hash
func (bh *BlockHandle) BlockSha() (btcwire.ShaHash, error) {
	return bh.Header.BlockSha(btcwire.ProtocolVersion)
}

// Save block writes a block to the db
func (chain *BlockChain) Save(header *btcwire.BlockHeader) (*BlockHandle, error) {
	parent, err := chain.Get(&header.PrevBlock)
	if err != nil {
		fmt.Errorf("No parent found for: %#v", header)
		return nil, err
	}
	// save a string to store our 'real' block with
	block_sha, err := header.BlockSha(btcwire.ProtocolVersion)
	if err != nil {
		fmt.Errorf("Something went wrong deriving the block hash for: %#v", header)
		return nil, err
	}
	bls_elems := []string{"blk", block_sha.String()}
	block_lookup_string := strings.Join(bls_elems, "")
	block := BlockHandle{Hash: block_sha,
		Header: *header,
		Depth:  parent.Depth + 1,
		Block:  block_lookup_string}
	var block_bytes bytes.Buffer
	encoder := gob.NewEncoder(&block_bytes)
	encoder.Encode(block)
	err = chain.Database.Put(block_sha.String(), block_bytes.Bytes())
	return &block, err
}

// fetch our handle to block through the databse
func (chain *BlockChain) Get(hash *btcwire.ShaHash) (*BlockHandle, error) {
	block_bytes, err := chain.Database.Get(hash.String())
	if err != nil {
		err := fmt.Errorf("No Block with Hash: %s", hash.String())
		return nil, err
	}
	block_reader := bytes.NewBuffer(block_bytes)
	decoder := gob.NewDecoder(block_reader)

	var block BlockHandle
	err = decoder.Decode(&block)
	if err != nil {
		log.Fatal("Error decoding block:", err)
	}

	return &block, nil
}

// find our blocks parent in the DB
func (chain *BlockChain) FindParent(bh *BlockHandle) (*BlockHandle, error) {
	parent_hash := bh.Header.PrevBlock
	parent, err := chain.Get(&parent_hash)
	if err != nil {
		return nil, err
	}
	return parent, nil
}

func InitializeBlockChain() (*BlockChain, error) {
	bc := BlockChain{Last: btcwire.GenesisHash, ChainHead: btcwire.GenesisHash, Genesis: btcwire.GenesisHash}
	db, err := gocask.NewGocask("bitcoin")
	bc.Database = db
	if err != nil {
		return nil, err
	}

	genesis := BlockHandle{Hash: btcwire.GenesisHash, Header: btcwire.GenesisBlock.Header}
	var block_bytes bytes.Buffer
	encoder := gob.NewEncoder(&block_bytes)
	encoder.Encode(genesis)
	err = bc.Database.Put(btcwire.GenesisHash.String(), block_bytes.Bytes())

	return &bc, nil
}

// create a block from a header we've read and add it to our blockchain
func (chain *BlockChain) InitializeBlock(header *btcwire.BlockHeader) (*BlockHandle, error) {
	block, err := chain.Save(header)
	if err != nil {
		err := fmt.Errorf("Error saving block: %s", err)
		return nil, err
	}
	// add a reference to our current chain head
	if chain.ChainHeadDepth < block.Depth {
		chain.ChainHead, _ = block.BlockSha()
		chain.ChainHeadDepth = block.Depth
	}
	return block, nil
}

func (chain *BlockChain) LongestPath() []*BlockHandle {
	var (
		parent *BlockHandle
		err    error
	)

	head, err := chain.Get(&chain.ChainHead)
	if err != nil {
		log.Fatal("No chainhead!")
	}
	parent, err = chain.FindParent(head)

	if err != nil {
		log.Fatal("No parent to root, something's fucky: %#v", head)
		return nil
	}
	var blocks []*BlockHandle
	blocks = append(blocks, head)

	for {
		blocks = append(blocks, parent)
		parent, err = chain.FindParent(parent)
		if err != nil {
			break // we're at the top
		}
	}
	return blocks
}

// create a block chain locator sequence to fetch our items
func (chain *BlockChain) CreateLocator() []*btcwire.ShaHash {
	longest_path := chain.LongestPath()

	var locator []*btcwire.ShaHash

	// A valid locator contains the last ten locator hashes
	for i, block := range longest_path {
		if i < 10 {
			block_sha, err := block.BlockSha()
			if err != nil {
				log.Fatal("Error building Block Hash: %#v", block)
			}
			locator = append(locator, &block_sha)
		} else {
			break
		}
	}

	if len(locator) < 10 {
		return locator
	}

	// as well as the hashes at position 10 + step ^ 2
	// incrementing step linearly, e.g 11, 14, 18, 26
	step := 1
	longest_path = longest_path[9:]
	for {
		if step >= len(longest_path) {
			break
		}
		block := longest_path[step]
		block_sha, err := block.BlockSha()
		if err != nil {
			log.Fatal("Error building Block Hash: %#v", block)
		}
		locator = append(locator, &block_sha)
		step = step * 2
	}
	return locator
}
