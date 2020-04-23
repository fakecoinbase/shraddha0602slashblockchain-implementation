package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/shraddha0602/blockchain-implementation/wallet"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes()
}

//create hash for transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]
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
	txin := TxInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(100, to)

	tx := Transaction{nil, []TxInput{txin}, []TxOutput{*txout}}
	tx.SetID()
	return &tx
}

//check if genesis block in transaction
func (tx *Transaction) IsCoinBase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

func NewTransactions(from, to string, amount int, UTXO *UTXOSet) *Transaction {
	var ins []TxInput
	var outs []TxOutput

	wallets, err := wallet.CreateWallets()
	Handle(err)
	w := wallets.GetWallet(from)
	publicKeyHash := wallet.PublicKeyHash(w.PublicKey)

	acc, validOuts := UTXO.FindSpendableOutput(publicKeyHash, amount)

	if acc < amount {
		log.Panic("Error : insufficient balance")
	}

	for txID, outputs := range validOuts {
		txid, err := hex.DecodeString(txID)
		Handle(err)

		for _, output := range outputs {
			input := TxInput{txid, output, nil, w.PublicKey}
			ins = append(ins, input)
		}
	}

	outs = append(outs, *NewTXOutput(amount, to))
	if acc > amount {
		outs = append(outs, *NewTXOutput(acc-amount, from))
	}

	tx := Transaction{nil, ins, outs}
	tx.ID = tx.Hash()
	UTXO.Blockchain.SignTransaction(&tx, w.PrivateKey)

	return &tx
}

//sign the transaction
func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTransac map[string]Transaction) {
	if tx.IsCoinBase() {
		return
	}

	for _, in := range tx.Inputs {
		if prevTransac[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("ERROR : Previous transaction doesn't exist")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inId, in := range txCopy.Inputs {
		prevTx := prevTransac[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTx.Outputs[in.Out].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID)
		Handle(err)
		sign := append(r.Bytes(), s.Bytes()...)

		tx.Inputs[inId].Signature = sign
	}

}

//creating transaction copy
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, in := range tx.Inputs {
		inputs = append(inputs, TxInput{in.ID, in.Out, nil, nil})
	}

	for _, out := range tx.Outputs {
		outputs = append(outputs, TxOutput{out.Value, out.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}
	return txCopy
}

//verify a transaction
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinBase() {
		return true
	}

	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("Previous transaction doesn't exists!!")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inId, in := range tx.Inputs {
		prevTx := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTx.Outputs[in.Out].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:(sigLen / 2)])
		s.SetBytes(in.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PubKey)
		x.SetBytes(in.PubKey[:(keyLen / 2)])
		y.SetBytes(in.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}
	return true
}

//convert transaction to string, for commandLine
func (tx Transaction) String() string {
	var transac []string

	transac = append(transac, fmt.Sprintf("-- Transaction %x", tx.ID))
	for i, input := range tx.Inputs {
		transac = append(transac, fmt.Sprintf("	Input %d : ", i))
		transac = append(transac, fmt.Sprintf("	 Transactiom ID : %x", input.ID))
		transac = append(transac, fmt.Sprintf("	 Output         : %d", input.Out))
		transac = append(transac, fmt.Sprintf("	 Signature      : %x", input.Signature))
		transac = append(transac, fmt.Sprintf("	 Public Key     : %x", input.PubKey))
	}

	for i, output := range tx.Outputs {
		transac = append(transac, fmt.Sprintf("  Output : %d", i))
		transac = append(transac, fmt.Sprintf("  Value  : %d", output.Value))
		transac = append(transac, fmt.Sprintf("  Script : %d", output.PubKeyHash))
	}
	return strings.Join(transac, "\n")
}
