package main

import (
	"fmt"
	"github.com/conformal/btcwire"
	"strings"
)

type BlockChain struct {
	Last           *Block
	Genesis        *Block
	ChainHead      *Block
	ChainHeadDepth int
	NodePointers   map[btcwire.ShaHash]*Block
}

type Block struct {
	Children     []*Block
	Parent       *Block
	Hash         btcwire.ShaHash
	Header       btcwire.BlockHeader
	Transactions []btcwire.MsgTx
	Depth        int
}

func InitializeBlockChain() *BlockChain {
	block := Block{Hash: btcwire.GenesisHash}
	bc := BlockChain{Last: &block, ChainHead: &block, Genesis: &block}
	bc.NodePointers = make(map[btcwire.ShaHash]*Block)
	bc.NodePointers[btcwire.GenesisHash] = &block
	return &bc
}

func (chain *BlockChain) LongestPath() []*Block {
	head := chain.ChainHead
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

func (chain *BlockChain) CreateLocator() []*btcwire.ShaHash {
	longest_path := chain.LongestPath()

	var locator []*btcwire.ShaHash

	for i, block := range longest_path {
		if i < 10 {
			locator = append(locator, &block.Hash)
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
		locator = append(locator, &block.Hash)
		step = step * 2
	}

	return locator
}

func (chain *BlockChain) AddBlock(header *btcwire.BlockHeader) (*Block, error) {

	header_sha, err := header.BlockSha(btcwire.ProtocolVersion)

	if err != nil {
		err := fmt.Errorf("Error building header blockhash: %s", err)
		return nil, err
	}

	parent := chain.NodePointers[header.PrevBlock]

	if parent == nil {
		err := fmt.Errorf("No parent with Header Hash: %s", header.PrevBlock.String())
		return nil, err
	}

	block := Block{Parent: parent, Hash: header_sha, Header: *header}
	chain.NodePointers[header_sha] = &block

	parent.Children = append(parent.Children, &block)
	block.Depth = parent.Depth + 1

	if chain.ChainHeadDepth < block.Depth {
		chain.ChainHead = &block
		chain.ChainHeadDepth = block.Depth

	}
	return &block, nil
}

func (block *Block) PrintSubtree(level int) {
	parent := block.Parent
	if parent == nil {
		fmt.Println("*")
	} else {

		if parent.Parent != nil && block != parent.Parent.Children[len(parent.Parent.Children)-1] {
			fmt.Println("|")
		}

		fmt.Printf("%s", strings.Repeat(string(' '), level-1))

		if parent != nil && block == parent.Children[len(parent.Children)-1] {
			fmt.Println("+")
		} else {
			fmt.Println("|")
		}
		fmt.Println("---")

		if len(block.Children) > 0 {
			fmt.Println("+")
		} else {
			fmt.Println(">")
		}
	}

	fmt.Printf("%s\n", block.Hash.String())

	for _, c := range block.Children {
		c.PrintSubtree(level + 1)
	}

}
