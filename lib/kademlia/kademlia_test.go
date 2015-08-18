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
	kc := kademliaCore{k}

	var contacts [100]Contact
	for i := 0; i < len(contacts); i++ {
		contacts[i] = Contact{NewRandomNodeID(), "127.0.0.1:8989"}
		if err := kc.ping(&pingRequest{RPCHeader{&contacts[i], k.NetworkID}},
			&pingResponse{}); err != nil {
			t.Errorf("Error on Ping %d: %s", i, err)
		}
	}

	args := findNodeRequest{RPCHeader{&contacts[0], k.NetworkID}, contacts[0].id}
	response := findNodeResponse{}
	if err := kc.findNode(&args, &response); err != nil {
		t.Errorf("Error on finding nodes: %s", err)
	}

	if len(response.contacts) != BucketSize {
		t.Errorf("Expected 'full' bucket of %d contacts: received %d", BucketSize, len(response.contacts))
	}
}

func TestStore(t *testing.T) {
	me := Contact{NewRandomNodeID(), "127.0.0.1:8989"}
	k := NewKademlia(&me, "test")
	kc := kademliaCore{k}
	someone := Contact{NewRandomNodeID(), "127.0.0.1:8989"}
	args := storeRequest{RPCHeader{&me, k.NetworkID}, "www.google.com", "A", net.ParseIP("74.125.224.72")}
	response := storeResponse{}

	if err := k.call(&someone, "kademliaCore.Store", &args, &response); err != nil {
		t.Errorf("Error storing www.google.com on remote node %s: %s", someone.String(), err)
	}

	if err := kc.store(&args, &response); err != nil {
		t.Errorf("Error storing www.google.com on local node %s: %s", me.String(), err)
	}
}

func TestIterativeFindNode(t *testing.T) {
	me := Contact{NewRandomNodeID(), "127.0.0.1:8989"}
	k := NewKademlia(&me, "test")
	kc := kademliaCore{k}

	var contacts [100]Contact
	for i := 0; i < len(contacts); i++ {
		contacts[i] = Contact{NewRandomNodeID(), "127.0.0.1:8989"}
		if err := kc.ping(&pingRequest{RPCHeader{&contacts[i], k.NetworkID}},
			&pingResponse{}); err != nil {
			t.Errorf("Error on Ping %d: %s", i, err)
		}
	}

	var contactRecords ContactRecList

	contactRecords = k.iterativeFindNode(contacts[0].id, 5)
	if len(contactRecords) > BucketSize {
		t.Errorf("Returned more than expected %d records: returned %d", BucketSize, len(contactRecords))
	}
}

func TestIterativeStore(t *testing.T) {
	me := Contact{NewRandomNodeID(), "127.0.0.1:8989"}
	k := NewKademlia(&me, "test")
	kc := kademliaCore{k}

	var contacts [100]Contact
	for i := 0; i < len(contacts); i++ {
		contacts[i] = Contact{NewRandomNodeID(), "127.0.0.1:8989"}
		if err := kc.ping(&pingRequest{RPCHeader{&contacts[i], k.NetworkID}},
			&pingResponse{}); err != nil {
			t.Errorf("Error on Ping %d: %s", i, err)
		}
	}

	k.iterativeStore("www.google.com", "A", net.ParseIP("74.125.224.72"))
}
