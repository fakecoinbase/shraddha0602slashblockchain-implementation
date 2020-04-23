package blockchain

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/dgraph-io/badger"
)

var (
	utxoPrefix   = []byte("utxo-")
	prefixLength = len(utxoPrefix)
)

type UTXOSet struct {
	Blockchain *Blockchain
}

func (utxo UTXOSet) CountTransacs() int {
	db := utxo.Blockchain.Database
	count := 0

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			count++
		}
		return nil
	})
	Handle(err)
	return count
}

func (u UTXOSet) Reindex() {
	db := u.Blockchain.Database

	u.DeleteByPrefix(utxoPrefix)

	UTXO := u.Blockchain.FindUnspentTransactions()

	err := db.Update(func(txn *badger.Txn) error {
		for txId, outs := range UTXO {
			key, err := hex.DecodeString(txId)
			Handle(err)
			key = append(utxoPrefix, key...)

			err = txn.Set(key, outs.SerializeOutputs())
			Handle(err)
		}
		return nil
	})
	Handle(err)
}

func (u *UTXOSet) Update(block *Block) {
	db := u.Blockchain.Database

	err := db.Update(func(txn *badger.Txn) error {
		for _, tx := range block.Transactions {
			if tx.IsCoinBase() == false {
				for _, in := range tx.Inputs {
					updatedOuts := TxOutputs{}
					inID := append(utxoPrefix, in.ID...)
					item, err := txn.Get(inID)
					Handle(err)
					v, err := item.Value()
					Handle(err)

					outs := DeserializeOutputs(v)

					for idx, out := range outs.Outputs {
						if idx != in.Out {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						if err := txn.Delete(inID); err != nil {
							log.Panic(err)
						}
					} else {
						if err := txn.Set(inID, updatedOuts.SerializeOutputs()); err != nil {
							log.Panic(err)
						}
					}
				}
			}
			newOutputs := TxOutputs{}
			for _, out := range tx.Outputs {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			txID := append(utxoPrefix, tx.ID...)
			if err := txn.Set(txID, newOutputs.SerializeOutputs()); err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
	Handle(err)
}

func (utxo *UTXOSet) DeleteByPrefix(prefix []byte) {
	deleteKeys := func(keysforDelete [][]byte) error {
		if err := utxo.Blockchain.Database.Update(func(txn *badger.Txn) error {
			for _, key := range keysforDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}
	collectSize := 100000
	utxo.Blockchain.Database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := it.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			if keysCollected == collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					log.Panic(err)
				}
				keysForDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}
		}
		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
}

//Finding all unspent transaction outputs
func (u UTXOSet) FindUTXOut(publicKeyHash []byte) []TxOutput {
	var UTXout []TxOutput
	db := u.Blockchain.Database

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			v, err := item.Value()
			Handle(err)
			outs := DeserializeOutputs(v)
			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(publicKeyHash) {
					UTXout = append(UTXout, out)
				}
			}
		}
		return nil
	})
	Handle(err)
	return UTXout
}

//for transactions that are not coin based
//find how many tokens available
func (u UTXOSet) FindSpendableOutput(publicKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutput := make(map[string][]int)
	db := u.Blockchain.Database
	accumulated := 0

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			key := item.Key()
			v, err := item.Value()
			Handle(err)
			key = bytes.TrimPrefix(key, utxoPrefix)
			txID := hex.EncodeToString(key)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(publicKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutput[txID] = append(unspentOutput[txID], outIdx)
				}
			}
		}
		return nil
	})
	Handle(err)

	return accumulated, unspentOutput
}
