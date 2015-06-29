package kademlia

import (
  "testing"
)

func TestRoutingTable(t *testing.T) {
  n1 := NewNodeID("FFFFFFFF00000000000000000000000000000000");
  n2 := NewNodeID("FFFFFFF000000000000000000000000000000000");
  n3 := NewNodeID("1111111100000000000000000000000000000000");
  rt := NewRoutingTable(&Contact{n1, "localhost:8000"})
  rt.Update(&Contact{n2, "localhost:8001"})
  rt.Update(&Contact{n3, "localhost:8002"})

  vec := rt.FindClosest(NewNodeID("2222222200000000000000000000000000000000"), 1)
  if len(vec) != 1 {
    t.Errorf("Returned incorrect number - %d closest nodes.  Expected 1.", len(vec))
    return
  }
  if !vec[0].node.id.Equals(n3) {
    t.Error(vec[0])
  }

  vec = rt.FindClosest(n2, 10)
  if len(vec) != 2 {
    t.Error(len(vec))
    return
  }
  if vec[0].node.id.Equals(n2) {
    t.Error(vec[0])
  }
  if vec[1].node.id.Equals(n3) {
    t.Error(vec[1])
  }
}
