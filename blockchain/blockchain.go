package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
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
func (chain *Blockchain) AddBlock(txs []*Transaction) {
	var lastHash []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.Value()
		return err
	})
	Handle(err)
	newBlock := CreateBlock(txs, lastHash)

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
func InitBlockchain(address string) *Blockchain {
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
		cbtx := CoinbaseTx(address, genesisData)
		genesis := GenesisBlock(cbtx)
		fmt.Println("Genesis created")

		err = txn.Set(genesis.Hash, genesis.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash)

		lasthash = genesis.Hash
		return err
	})
	Handle(err)
	blockchain := Blockchain{lasthash, db}
	return &blockchain
}

//if blockchain already exists
func ContinueBlockchain(address string) *Blockchain {
	if DBexists() == false {
		fmt.Println("No existsing database found, create one")
		runtime.Goexit()
	}
	var lastHash []byte

	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.Value()
		return err
	})
	Handle(err)
	chain := Blockchain{lastHash, db}
	return &chain

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

func (chain *Blockchain) FindUnspentTransactions(publicKeyHash []byte) []Transaction {
	var unspent []Transaction
	spent := make(map[string][]int)

	itr := chain.Iterator()
	for {
		block := itr.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spent[txID] != nil {
					for _, spentOut := range spent[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.IsLockedWithKey(publicKeyHash) {
					unspent = append(unspent, *tx)
				}
			}

			if tx.IsCoinBase() == false {
				for _, in := range tx.Inputs {
					if in.UsesKey(publicKeyHash) {
						inTxID := hex.EncodeToString(in.ID)
						spent[inTxID] = append(spent[inTxID], in.Out)
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return unspent
}

//Finding all unspent transaction outputs
func (chain *Blockchain) FindUTXOut(publicKeyHash []byte) []TxOutput {
	var UTXout []TxOutput
	unspentTransac := chain.FindUnspentTransactions(publicKeyHash)

	for _, tx := range unspentTransac {
		for _, out := range tx.Outputs {
			if out.IsLockedWithKey(publicKeyHash) {
				UTXout = append(UTXout, out)
			}
		}
	}

	return UTXout
}

//for transactions that are not coin based
//find how many tokens available
func (chain *Blockchain) FindSpendableOutput(publicKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutput := make(map[string][]int)
	unspentTransac := chain.FindUnspentTransactions(publicKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTransac {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Outputs {
			if out.IsLockedWithKey(publicKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutput[txID] = append(unspentOutput[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutput
}

//finding transaction
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	itr := bc.Iterator()

	for {
		block := itr.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("Transaction doesn't exists!!")
}

//signing the transaction
func (bc *Blockchain) SignTransaction(tx *Transaction, privateKey ecdsa.PrivateKey) {
	prevTransacs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTransac, err := bc.FindTransaction(in.ID)
		Handle(err)
		prevTransacs[hex.EncodeToString(prevTransac.ID)] = prevTransac
	}

	tx.Sign(privateKey, prevTransacs)
}

//verifying a transaction
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	prevTransacs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTransac, err := bc.FindTransaction(in.ID)
		Handle(err)
		prevTransacs[hex.EncodeToString(prevTransac.ID)] = prevTransac
	}
	return tx.Verify(prevTransacs)
}
