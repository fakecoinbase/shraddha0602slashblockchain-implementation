package blockchain

import (
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

// Database Path
const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST"
	genesisData = "First Transaction from Genesis"
)

type Blockchain struct {
	LastHash []byte
	Database *badger.DB
}

// To implement feature to iterate through blockchain and access each Block
type BlockchainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

//to check if database exists
func DBexists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// Adds block to Blockchain
func (chain *Blockchain) AddBlock(data string) {
	var lastHash []byte
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.Value()
		return err
	})
	Handle(err)
	newBlock := CreateBlock(data, lastHash)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)
		chain.LastHash = newBlock.Hash
		return err
	})
	Handle(err)
}

//Initialize Blockchain on start
func InitBlockchain() *Blockchain {
	var lasthash []byte

	if DBexists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			// If no existing blockchain found
			fmt.Println("No existing blockchain found ")
			genesis := GenesisBlock()
			fmt.Println("Genesis Proved")

			err = txn.Set(genesis.Hash, genesis.Serialize())
			Handle(err)
			err = txn.Set([]byte("lh"), genesis.Hash)

			lasthash = genesis.Hash
			return err
		} else {
			// if blockchain exists
			item, err := txn.Get([]byte("lh"))
			Handle(err)
			lasthash, err = item.Value()
			return err
		}
	})
	Handle(err)
	blockchain := Blockchain{lasthash, db}
	return &blockchain
}

// function to convert Blockchain to BlockchainIterator
func (chain *Blockchain) Iterator() *BlockchainIterator {
	itr := &BlockchainIterator{chain.LastHash, chain.Database}
	return itr
}

//iterate backwords using previous Hash stored in db
func (itr *BlockchainIterator) Next() *Block {
	var block *Block

	err := itr.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(itr.CurrentHash)
		encodedBlock, err := item.Value()
		block = Deserialize(encodedBlock)
		return err
	})
	Handle(err)
	itr.CurrentHash = block.PrevHash
	return block
}
