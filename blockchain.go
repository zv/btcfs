package main 

import (
	"fmt"
	"github.com/conformal/btcwire"
  "encoding/binary"
  "bytes"
)

type BlockChain struct {
	Last *Block
	ChainHead *Block 
	ChainHeadDepth int
	NodePointers map[btcwire.ShaHash] *Block 
}

type Block struct {
	Children []*Block
	Parent *Block
	MerkleRoot btcwire.ShaHash
  Hash btcwire.ShaHash 
  PrevBlock btcwire.ShaHash 
	Height int
}

func InitializeBlockChain() *BlockChain {
	block := Block{MerkleRoot: btcwire.GenesisHash}
  bc := BlockChain{Last: &block, ChainHead: &block} 
	bc.NodePointers = make(map[btcwire.ShaHash] *Block)	
	bc.NodePointers[btcwire.GenesisHash] = &block 
	return &bc	
}

/*
func (node *Block) LongestPath() int, *Block {
	var children_height map[*Block] int 

	if len(node.Children) == 0 {
		return 0
	}	 

	var path_lengths []int

	for i, child := range subtree.Children {
		child_path_length, _ := child.longest_path()
		children_height[&child] = child.longest_path()
	}

	largest := 0
	var most_distant_node *Block
	for k, v := range children_height {
		if v > largest {
			largest := v 
			most_distant_node = k
		}
	}	
	
	return largest, most_distant_node	
} 
*/
func (chain *BlockChain) LongestPath() []*Block {
	head   := chain.ChainHead
	parent := head.Parent	

  var blocks []*Block  

	blocks = append(blocks, head)

	for {
		if parent == nil {
			break	
		}
		blocks = append(blocks, parent)
		parent = parent.Parent 
	}
	return blocks
} 

func (chain *BlockChain) CreateLocator() []btcwire.ShaHash {
	longest_path := chain.LongestPath()
	var locator []btcwire.ShaHash

	for i, block := range longest_path {
		if i > 10 {
			locator = append(locator, block.MerkleRoot)
		} else {
			break
		}		
	}			

	if len(locator) < 10 {
		return locator 
	}

	step := 1
	longest_path = longest_path[9:]
	for {
		if step >= len(longest_path) {
			break			
		} 
		block := longest_path[step]
		locator = append(locator, block.MerkleRoot)
		step = step * 2 
	}
	return locator 
} 

func (chain *BlockChain) AddBlock(header *btcwire.BlockHeader) (*Block, error) {

	buf := new(bytes.Buffer)
  binary.Write(buf, binary.LittleEndian, header.Version) 
  binary.Write(buf, binary.LittleEndian, header.PrevBlock) 
  binary.Write(buf, binary.LittleEndian, header.MerkleRoot) 
  binary.Write(buf, binary.LittleEndian, header.Timestamp) 
  binary.Write(buf, binary.LittleEndian, header.Bits) 
  binary.Write(buf, binary.LittleEndian, header.Nonce) 
  
  header_hash := btcwire.DoubleSha256(buf.Bytes())
  header_sha, _ := btcwire.NewShaHash(header_hash)
  
  fmt.Printf("header_sha: %s", header_sha.String())

	parent := chain.NodePointers[*header_sha]

	if parent == nil {
		err := fmt.Errorf("No parent with Header Hash: %s", header_sha.String())	
		return nil, err 
	} 

	block := Block{Parent: parent, Hash: *header_sha, PrevBlock: header.PrevBlock}
	chain.NodePointers[*header_sha] = &block 

	parent.Children = append(parent.Children, &block)  
	block.Height    = parent.Height + 1

	if chain.ChainHeadDepth < block.Height {
		chain.ChainHead = &block 
		chain.ChainHeadDepth = block.Height 

	}
	return &block, nil
}



