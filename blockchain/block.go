package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
	Nonce    int
}

//Create a new block
func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), prevHash, 0}
	pow := Proof(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

// Create Genesis Block (very first block)
func GenesisBlock() *Block {
	return CreateBlock("Genesis", []byte{})
}

// Serialize data as bytes to input badgerDB
// @Params (Block struct)
// @return ([]byte)
func (block *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(block)
	Handle(err)
	return res.Bytes()
}

//Deserialze data from []byte to *Block
//@Params ([]byte)
//@return (*Block)
func Deserialize(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)
	Handle(err)
	return &block
}

//function to Handle error
func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
