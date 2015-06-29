package kademlia

import (
  "testing"
  "fmt"
)

func TestNodeID(t *testing.T) {
  a := NodeID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19};
  b := NodeID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 19, 18};
  c := NodeID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,  0,  0,  0,  0,  0,  0,  0,  1,  1};

  if !a.Equals(a) {
    t.Fail()
  }
  if a.Equals(b) {
    t.Fail()
  }

  if !a.Xor(b).Equals(c) {
    t.Error(a.Xor(b))
  }

  if c.PrefixLen() != 151 {
    t.Error(c.PrefixLen())
  }

  if b.Less(a) {
    t.Fail()
  }

  str_id := "0123456789abcdef0123456789abcdef01234567";
  if NewNodeID(str_id).String() != str_id {
    t.Error(NewNodeID(str_id).String());
  }
}

type Foo struct {
  a int
}

func (f Foo) String() string {
  return "foo"
}

type FmtWriter struct {}

func (fw FmtWriter) Write(p []byte) (n int, err error) {
  fmt.Printf("gob(%d): %x\n", len(p), p)
  return len(p), nil
}
