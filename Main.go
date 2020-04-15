package main

import (
	"os"

	"github.com/shraddha0602/blockchain-implementation/cli"
)

func main() {
	defer os.Exit(0)
	cli := cli.CommandLine{}
	cli.Run()

	//w := wallet.MakeWallet()
	//w.Address()
}
