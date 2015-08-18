package kademlia

import (
	"fmt"
)

// Contact represents a node id and address record in the DHT
type Contact struct {
	id      NodeID
	address string
}

func (contact *Contact) String() string {
	return fmt.Sprintf("Contact(\"%s\", \"%s\")", contact.id, contact.address)
}

// Less compares contacts by id to determine which is lower.
func (contact *Contact) Less(other interface{}) bool {
	return contact.id.Less(other.(*Contact).id)
}
