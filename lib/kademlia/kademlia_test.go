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
  kc := KademliaCore{k}
  someone := Contact{NewRandomNodeID(), "127.0.0.1:8989"}
  args := StoreRequest{RPCHeader{&me, k.NetworkID}, "www.google.com", "A", net.ParseIP("74.125.224.72")}
  response := StoreResponse{}

  if err := k.Call(&someone, "KademliaCore.Store", &args, &response); err != nil {
    t.Errorf("Error storing www.google.com on remote node %s: %s", someone.String(), err)
  }

  if err := kc.Store(&args, &response); err != nil {
    t.Errorf("Error storing www.google.com on local node %s: %s", me.String(), err)
  }
}

func TestIterativeFindNode(t *testing.T) {
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

  var contact_records ContactRecList

  contact_records = k.IterativeFindNode(contacts[0].id, 5)
  if len(contact_records) > BucketSize {
    t.Errorf("Returned more than expected %d records: returned %d", BucketSize, len(contact_records))
  }
}

func TestIterativeStore(t *testing.T) {
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

  k.IterativeStore("www.google.com", "A", net.ParseIP("74.125.224.72"))
}
