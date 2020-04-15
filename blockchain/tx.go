package blockchain

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

//unlock data in Inputs
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

//unlock data in Outputs
func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}
