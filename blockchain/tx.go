package blockchain

import (
	"bytes"
	"encoding/gob"

	"github.com/shraddha0602/blockchain-implementation/wallet"
)

type TxOutput struct {
	Value      int    //Value in tokens
	PubKeyHash []byte // to unlock tokens in Value
}

//references to prev output
type TxInput struct {
	ID        []byte //transaction ID
	Out       int    //index of transaction
	Signature []byte
	PubKey    []byte
}

type TxOutputs struct {
	Outputs []TxOutput
}

func NewTXOutput(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}

//unlock input
func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

//lock the output
func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := wallet.Base58Decode(address)
	// remove version and checksum
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

//checks to see if the o/p is locked with Public Key
func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

//Serialize Outputs
func (outputs TxOutputs) SerializeOutputs() []byte {
	var buffer bytes.Buffer
	encode := gob.NewEncoder(&buffer)
	err := encode.Encode(outputs)
	Handle(err)
	return buffer.Bytes()
}

//Deserialize Outputs
func DeserializeOutputs(outputs []byte) TxOutputs {
	var outs TxOutputs
	decode := gob.NewDecoder(bytes.NewReader(outputs))
	err := decode.Decode(&outs)
	Handle(err)
	return outs
}
