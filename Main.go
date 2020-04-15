package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/shraddha0602/blockchain-implementation/blockchain"
)

type CommandLine struct{}

//command line description
func (cli *CommandLine) printUsage() {
	fmt.Println("Usage : ")
	fmt.Println("getbalance -address <ADDRESS> - get the balance for given adress")
	fmt.Println(" createblockchain - address <ADDRESS> - creates a blockchain")
	fmt.Println("print - prints the blockchain")
	fmt.Println(" send -from <FROM> -to <TO> -amount <AMOUNT> -Send amount")
}

//func to Validate arguments input through command line
func (cli *CommandLine) ValidateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		//@runtime.Goexit() - ensure shutting of Go routine & not corrupting data in db
		runtime.Goexit()
	}
}

//func to handle cli print blockchain
func (cli *CommandLine) printBlockchain() {
	chain := blockchain.ContinueBlockchain("")
	defer chain.Database.Close()
	itr := chain.Iterator()
	for {
		block := itr.Next()
		fmt.Printf("Previous Hash : %x\n", block.PrevHash)
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

//create the blockchain
func (cli *CommandLine) createBlockchain(address string) {
	chain := blockchain.InitBlockchain(address)
	chain.Database.Close()
	fmt.Println("\nBlockchain Created!!")
}

// Get all unspent transac and get balance
func (cli *CommandLine) getBalance(address string) {
	chain := blockchain.ContinueBlockchain(address)
	defer chain.Database.Close()

	bal := 0
	UTXouts := chain.FindUTXOut(address)

	for _, out := range UTXouts {
		bal += out.Value
	}

	fmt.Printf("Balance of account %s is %d\n", address, bal)
}

//send tokens from one acct to other
func (cli *CommandLine) send(from, to string, amt int) {
	chain := blockchain.ContinueBlockchain(from)
	defer chain.Database.Close()

	tx := blockchain.NewTransactions(from, to, amt, chain)
	chain.AddBlock([]*blockchain.Transaction{tx})
	fmt.Println("\nTransaction successful!!")
}

// run cli commands
func (cli *CommandLine) run() {
	cli.ValidateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printCmd := flag.NewFlagSet("print", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address of account")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to create Blockchain")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmt := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "send":
		err := sendCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "print":
		err := printCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	default:
		cli.printUsage()
		runtime.Goexit()
	}

	//Parse if no error thrown returns bool
	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmt <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}
		cli.send(*sendFrom, *sendTo, *sendAmt)
	}

	if printCmd.Parsed() {
		cli.printBlockchain()
	}
}

func main() {
	defer os.Exit(0)
	cli := CommandLine{}
	cli.run()
}
