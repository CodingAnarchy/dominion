package kademlia

import (
  "testing"
  "fmt"
  "strings"
)

func rightPad2Len(s string, pad string, overallLen int) string {
  padCount := 1 + ((overallLen - len(pad))/len(pad))
  return (s + strings.Repeat(pad, padCount))[:overallLen]
}

func TestNodeID(t *testing.T) {
  a := NodeID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19};
  b := NodeID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 19, 18};
  c := NodeID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,  0,  0,  0,  0,  0,  0,  0,  1,  1};

  if !a.Equals(a) {
    t.Errorf("%s not equal to itself!\n", a)
  }
  if a.Equals(b) {
    t.Errorf("%s equal to %s\n", a, b)
  }

  if !a.Xor(b).Equals(c) {
    t.Errorf("%s should equal %s\n", a.Xor(b), c)
  }

  if c.PrefixLen() != 151 {
    t.Errorf("Expected prefix length of 151: obtained %d", c.PrefixLen())
  }

  if b.Less(a) {
    t.Errorf("Expected %s to not be less than %s", b, a)
  }

  str_id := "0123456789abcdef0123456789abcdef01234567"
  if NewNodeID(str_id).String() != str_id {
    t.Errorf("Did not properly translate as NodeID and return %s: obtained %s", str_id, NewNodeID(str_id).String());
  }

  domain_node := fmt.Sprintf("%x", "www.google.com")
  if NewNodeID(domain_node).String() != rightPad2Len(domain_node, "0", 40) {
    t.Errorf("Did not properly translate as NodeID and return %s: obtained %s",
      rightPad2Len(domain_node, "0", 40), NewNodeID(domain_node).String());
  }
}
