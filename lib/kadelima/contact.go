package kademlia

import (
  "fmt"
)

type Contact struct {
  id NodeID
  address string
}

func (contact *Contact) String() string {
  return fmt.Sprintf("Contact(\"%s\", \"%s\")", contact.id, contact.address)
}
