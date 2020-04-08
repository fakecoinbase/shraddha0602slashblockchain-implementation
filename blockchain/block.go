package main

import (
	"fmt"

	github.com/shraddha0602/blockchain-implementation/blockchain"
)

func main() {
	chain := blockchain.InitBlockchain()

	chain.AddBlock("First Block after Genesis")
	chain.AddBlock("Second Block after Genesis")
	chain.AddBlock("Third Block after Genesis")

	for _, block := range chain.Blocks {
		fmt.Printf("Previous Hash : %x\n", block.PrevHash)
		fmt.Printf("Data : %s \n", block.Data)
		fmt.Printf("Hash : %x\n", block.Hash)

		fmt.Printf("Hello there")
	}
}
