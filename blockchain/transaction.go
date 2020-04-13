package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxOutput struct {
	Value  int    //Value in tokens
	PubKey string // to unlock tokens in Value
}

//references to prev output
type TxInput struct {
	ID  []byte //transaction ID
	Out int    //index of transaction
	Sig string
}

//create hash of Input of a transaction
func (tx *Transaction) SetID() {
	var buff bytes.Buffer
	var hash [32]byte

	encode := gob.NewEncoder(&buff)
	err := encode.Encode(tx)
	Handle(err)

	hash = sha256.Sum256(buff.Bytes())
	tx.ID = hash[:]
}

//transaction if genesis block
func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}
	txin := TxInput{[]byte{}, -1, data}
	txout := TxOutput{100, to}

	tx := Transaction{nil, []TxInput{txin}, []TxOutput{txout}}
	tx.SetID()
	return &tx
}

//check if genesis block in transaction
func (tx *Transaction) IsCoinBase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

//unlock data in Inputs
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

//unlock data in Outputs
func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}
