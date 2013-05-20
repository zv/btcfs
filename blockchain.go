package main 

import (
	"fmt"
	"github.com/conformal/btcwire"
)

type BlockChain struct {
	Last *Block
	ChainHead *Block 
	ChainHeadDepth int
	NodePointers map[btcwire.Shahash] *Block 
}

type Block struct {
	Children []*Block
	Parent *Block
	MerkleRoot btcwire.Shahash
	Height int
}

func InitializeBlockChain() *BlockChain {
	block := Block{MerkleRoot: *btcwire.GenesisMerkleRoot}
	bc.NodePointers = make(map[btcwire.Shahash] block)	
	bc.NodePointers[block.MerkleRoot] = *block 
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
	blocks := append(nil, head)
	for {
		if parent == nil {
			break	
		}
		blocks = append(blocks, parent)
		parent = parent.Parent 
	}
	return blocks
} 

func (chain *BlockChain) CreateLocator() []*btcwire.Shahash {
	longest_path := chain.LongestPath()
	var locator []*Block
	for i, block := range longest_path {
		if i > 10 {
			locator := append(locator, block)
		} else {
			break
		}		
	}			

	if len(longest_path) > 10 {
		return longest_path
	}

	step := 1
	longest_path = longest_path[9:]
	for {
		if step >= len(longest_path) {
			break			
		} 
		block := longest_path[step]
		locator := append(locator, block)
		step = step * 2 
	}
	return longest_path
} 

func (chain *BlockChain) AddBlock(header *btcwire.BlockHeader) (*Block, error) {
	parent := chain.NodePointers[header.PrevBlock]
	if parent == nil {
		err := fmt.Errorf("No parent with Merkle root %s", header.PrevBlock.String())	
		return nil, err 
	} 
	block := Block{Parent: parent, MerkleRoot: header.MerkleRoot}
	chain.NodePointers[header.MerkleRoot] = block 

	parent.Children = append(parent.Children, block)  
	block.Height    = node.Height + 1

	if chain.ChainHeadSize < node.Height {
		chain.ChainHead = block 
	}
	return &block, nil
}



