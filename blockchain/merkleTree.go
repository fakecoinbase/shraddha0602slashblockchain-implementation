package blockchain

import (
	"crypto/sha256"
)

type MerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := MerkleNode{}

	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		node.Data = hash[:]
	} else {
		temp := append(left.Data, right.Data...)
		hash := sha256.Sum256(temp)
		node.Data = hash[:]
	}
	node.Left = left
	node.Right = right
	return &node
}

func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	if len(nodes)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, temp := range data {
		node := NewMerkleNode(nil, nil, temp)
		nodes = append(nodes, *node)
	}

	for i := 0; i < len(data)/2; i++ {
		var lvl []MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			lvl = append(lvl, *node)
		}
		nodes = lvl
	}
	tree := MerkleTree{&nodes[0]}
	return &tree
}
