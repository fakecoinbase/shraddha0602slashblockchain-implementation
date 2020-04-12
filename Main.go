package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/shraddha0602/blockchain-implementation/blockchain"
)

type CommandLine struct {
	blockchain *blockchain.Blockchain
}

//command line description
func (cli *CommandLine) printUsage() {
	fmt.Println("Usage : ")
	fmt.Println("add -block <BLOCK_DATA> - adds block to the blockchain")
	fmt.Println("print - prints the blockchain")
}

//func to Validate arguments input through command line
func (cli *CommandLine) ValidateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		//@runtime.Goexit() - ensure shutting of Go routine & not corrupting data in db
		runtime.Goexit()
	}
}

//func to handle cli command to add block
func (cli *CommandLine) addBlock(data string) {
	cli.blockchain.AddBlock(data)
	fmt.Println("Block Added!!")
}

//func to handle cli print blockchain
func (cli *CommandLine) printBlockchain() {
	itr := cli.blockchain.Iterator()
	for {
		block := itr.Next()
		fmt.Printf("Previous Hash : %x\n", block.PrevHash)
		fmt.Printf("Data : %s \n", block.Data)
		fmt.Printf("Hash : %x\n", block.Hash)

		pow := blockchain.Proof(block)
		fmt.Printf("PoW : %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		//check if block is Genesis block
		//genesis block has no prev hash, hence len = 0
		if len(block.PrevHash) == 0 {
			break
		}
	}
}

// run cli commands
func (cli *CommandLine) run() {
	cli.ValidateArgs()

	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "Block data")

	switch os.Args[1] {
	case "add":
		err := addBlockCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "print":
		err := printCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	default:
		cli.printUsage()
		runtime.Goexit()
	}

	//Parse if no error thrown returns bool
	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			runtime.Goexit()
		}
		cli.addBlock(*addBlockData)
	}
	if printCmd.Parsed() {
		cli.printBlockchain()
	}
}

func main() {
	defer os.Exit(0)
	chain := blockchain.InitBlockchain()
	//to ensure Db closes securely
	//defer only runs when proper exit of Go routines
	//so we use @Goexit()
	defer chain.Database.Close()
	cli := CommandLine{chain}
	cli.run()
}
