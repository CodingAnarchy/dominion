package kademlia

import "testing"

func TestContact(t *testing.T) {
  a := &Contact{NewNodeID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"}
  b := &Contact{NewNodeID("1111111100000000000000000000000000000000"), "localhost:8001"}

  if !b.Less(a) {
    t.Errorf("Expected %s to be less than %s", b, a)
  }
}
