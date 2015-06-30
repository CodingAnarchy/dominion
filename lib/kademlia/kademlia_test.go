package kademlia

import (
  "net"
  "testing"
)

func TestPing(t *testing.T) {
  me := Contact{NewRandomNodeID(), "127.0.0.1:8989"}
  k := NewKademlia(&me, "test")
  k.Serve()

  someone := Contact{NewRandomNodeID(), "127.0.0.1:8989"}
  if err := k.sendPingQuery(&someone); err != nil {
    t.Error(err)
  }
}

func TestFindNode(t *testing.T) {
  me := Contact{NewRandomNodeID(), "127.0.0.1:8989"}
  k := NewKademlia(&me, "test")
  kc := KademliaCore{k}

  var contacts [100]Contact
  for i := 0; i < len(contacts); i++ {
    contacts[i] = Contact{NewRandomNodeID(), "127.0.0.1:8989"}
    if err := kc.Ping(&PingRequest{RPCHeader{&contacts[i], k.NetworkID}},
                      &PingResponse{}); err != nil {
                      t.Error(err)
    }
  }

  args := FindNodeRequest{RPCHeader{&contacts[0], k.NetworkID}, contacts[0].id}
  response := FindNodeResponse{}
  if err := kc.FindNode(&args, &response); err != nil {
    t.Error(err)
  }

  if len(response.contacts) != BucketSize {
    t.Fail()
  }
}

func TestStore(t *testing.T) {
  me := Contact{NewRandomNodeID(), "127.0.0.1:8989"}
  k := NewKademlia(&me, "test")
  someone := Contact{NewRandomNodeID(), "127.0.0.1:8989"}
  args := StoreRequest{RPCHeader{&someone, k.NetworkID}, "www.google.com", "A", net.ParseIP("74.125.224.72")}
  response := StoreResponse{}

  if err := k.Call(&someone, "KademliaCore.Store", &args, &response); err != nil {
    t.Errorf("Error storing www.google.com on remote node %s: %s", someone, err)
  }
}
