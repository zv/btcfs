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
		log.Fatal("No parent found for: %#v", header)
		return nil, err
	}
	// save a string to store our 'real' block with
	block_sha, err := header.BlockSha(btcwire.ProtocolVersion)
	if err != nil {
		log.Fatal("Something went wrong deriving the block hash for: %#v", header)
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
	gob.Register(btcwire.ShaHash{})
	db, err := gocask.NewGocask("bitcoin")
  chain_bytes, err := db.Get("chainhead")
  if err == nil {
    var bc BlockChain 
    bc.Database = db
    head_hash, err := btcwire.NewShaHash(chain_bytes)
    if err != nil {
      return nil, err
    }
    bc.Last = *head_hash
    bc.ChainHead = *head_hash
    chainhead_handle, err := bc.Get(head_hash)
    if err != nil {
      log.Fatal("Couldn't derive Chain Head")
      return nil, nil
    }

    bc.ChainHeadDepth = chainhead_handle.Depth 
    bc.Genesis = btcwire.GenesisHash
    fmt.Printf("Loaded previous  blockchain with a depth of: %d\n", bc.ChainHeadDepth)
    return &bc, nil
  } else {
    bc := BlockChain{Last: btcwire.GenesisHash, ChainHead: btcwire.GenesisHash, Genesis: btcwire.GenesisHash}
    err = db.Put("chainhead", btcwire.GenesisHash.Bytes())
    if err != nil {
      return nil, err
    }
    bc.Database = db

    // save a string to store our 'real' block with
    block_sha, err := btcwire.GenesisBlock.Header.BlockSha(btcwire.ProtocolVersion)
    if err != nil {
      log.Fatal("Something went wrong deriving the genesis block hash")
      return nil, err
    }
    bls_elems := []string{"blk", block_sha.String()}
    block_lookup_string := strings.Join(bls_elems, "")
    block := BlockHandle{Hash: block_sha,
      Header: btcwire.GenesisBlock.Header,
      Depth:  0, 
      Block:  block_lookup_string}
    var block_bytes bytes.Buffer
    encoder := gob.NewEncoder(&block_bytes)
    encoder.Encode(block)
    err = db.Put(block_sha.String(), block_bytes.Bytes())
    if err != nil {
      log.Fatal("Couldn't save Genesis Block Header")
      return nil, err
    }

    // genesis := BlockHandle{Hash: btcwire.GenesisHash, Header: btcwire.GenesisBlock.Header}
    return &bc, nil

  }
  return nil, nil  
}

// create a block from a header we've read and add it to our blockchain
func (chain *BlockChain) InitializeBlock(header *btcwire.BlockHeader) (*BlockHandle, error) {
	block, err := chain.Save(header)
	if err != nil {
		err := fmt.Errorf("Error saving block: %s", err)
		return nil, err
	}
	// add a reference to our current chain head
  chain_head_bytes, err := chain.Database.Get("chainhead")
  if err != nil {
    return nil, err
  }

  chain_head_hash, err := btcwire.NewShaHash(chain_head_bytes)
  if err != nil {
    return nil, err
  }

  chain_head, err := chain.Get(chain_head_hash)
  if err != nil {
    return nil, err
  }

	if chain_head.Depth < block.Depth {
    chain_head_sha, _ := block.BlockSha()
    chain.ChainHead = chain_head_sha 
    // we should defer this to only write when we're about to exit because of the
    // append only nature of bitcask 
    chain.Database.Put("chainhead", chain_head_sha.Bytes()) 
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
