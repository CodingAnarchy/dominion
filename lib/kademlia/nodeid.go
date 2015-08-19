package kademlia

import (
	"encoding/hex"
	"math/rand"
)

const idLength = 20

// NodeID type for storing node identifier
type NodeID [idLength]byte

// NewNodeID creates a new node id based on string input
func NewNodeID(data string) (ret NodeID) {
	decoded, _ := hex.DecodeString(data)
	length := idLength
	if len(decoded) < idLength {
		length = len(decoded)
	}
	for i := 0; i < length; i++ {
		ret[i] = decoded[i]
	}
	return
}

// NewRandomNodeID creates a new random node ID
func NewRandomNodeID() (ret NodeID) {
	for i := 0; i < idLength; i++ {
		ret[i] = uint8(rand.Intn(256))
	}
	return
}

func (node NodeID) String() string {
	return hex.EncodeToString(node[0:idLength])
}

// Equals - test equality between node IDs
func (node NodeID) Equals(other NodeID) bool {
	for i := 0; i < idLength; i++ {
		if node[i] != other[i] {
			return false
		}
	}
	return true
}

// Xor - apply exclusive or between two NodeIDs
func (node NodeID) Xor(other NodeID) (ret NodeID) {
	for i := 0; i < idLength; i++ {
		ret[i] = node[i] ^ other[i]
	}
	return
}

// PrefixLen - calculate and return prefix length of the node
func (node NodeID) PrefixLen() (ret int) {
	for i := 0; i < idLength; i++ {
		for j := 0; j < 8; j++ {
			if (node[i]>>uint8(7-j))&0x1 != 0 {
				return i*8 + j
			}
		}
	}
	return idLength*8 - 1
}

// Less -compare Node IDs to determine which is lower
func (node NodeID) Less(other interface{}) bool {
	for i := 0; i < idLength; i++ {
		if node[i] != other.(NodeID)[i] {
			return node[i] < other.(NodeID)[i]
		}
	}
	return false
}
