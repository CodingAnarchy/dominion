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
    t.Errorf("Error on sending ping query: %s", err)
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
                      t.Errorf("Error on Ping %d: %s", i, err)
    }
  }

  // Repeat test of ping to contacts within list to test
  // update of node that already exists
  for i, contact := range contacts {
    if err := kc.Ping(&PingRequest{RPCHeader{&contact, k.NetworkID}},
                      &PingResponse{}); err != nil {
                      t.Errorf("Error testing repeat Ping %d: %s", i, err)
    }
  }

  args := FindNodeRequest{RPCHeader{&contacts[0], k.NetworkID}, contacts[0].id}
  response := FindNodeResponse{}
  if err := kc.FindNode(&args, &response); err != nil {
    t.Errorf("Error on finding nodes: %s", err)
  }

  if len(response.contacts) != BucketSize {
    t.Errorf("Expected 'full' bucket of %d contacts: received %d", BucketSize, len(response.contacts))
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
