package kademlia

import (
	"container/heap"
	"container/list"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sort"
)

// Core kademlia structs

type kademlia struct {
	routes    *RoutingTable
	NetworkID string
	domains   *DomainStore
}

type kademliaCore struct {
	kad *kademlia
}

// RPC Request and Response structs

type RPCHeader struct {
	Sender    *Contact
	NetworkID string
}

type pingRequest struct {
	RPCHeader
}

type pingResponse struct {
	RPCHeader
}

type storeRequest struct {
	RPCHeader
	domain string
	typ    string
	ip     net.IP
}

type storeResponse struct {
	RPCHeader
}

type findNodeRequest struct {
	RPCHeader
	target NodeID
}

type findNodeResponse struct {
	RPCHeader
	contacts []Contact
}

type findValueRequest struct {
	RPCHeader
	domain string
	typ    string
}

type findValueResponse struct {
	RPCHeader
	ip       net.IP
	contacts []Contact
}

// Data structures for internal use

// ContactHeap
type ContactHeap []Contact

func (c ContactHeap) Len() int           { return len(c) }
func (c ContactHeap) Less(i, j int) bool { return c[i].Less(&c[j]) }
func (c ContactHeap) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

func (c *ContactHeap) Push(x interface{}) {
	*c = append(*c, x.(Contact))
}

func (c *ContactHeap) Pop() interface{} {
	old := *c
	n := len(old)
	x := old[n-1]
	*c = old[0 : n-1]
	return x
}

// ContactRecList
type ContactRecList []*ContactRecord

func (cr ContactRecList) Len() int           { return len(cr) }
func (cr ContactRecList) Less(i, j int) bool { return cr[i].Less(cr[j]) }
func (cr ContactRecList) Swap(i, j int)      { cr[i], cr[j] = cr[j], cr[i] }

// kademlia functionality

func NewKademlia(self *Contact, networkID string) (ret *kademlia) {
	ret = new(kademlia)
	ret.routes = NewRoutingTable(self)
	ret.NetworkID = networkID
	ret.domains = NewDomainStore()
	return
}

func (k *kademlia) Update(contact *Contact, table *RoutingTable) {
	prefix_length := contact.id.Xor(table.node.id).PrefixLen()
	bucket := table.buckets[prefix_length]
	var elt *list.Element
	for elt = bucket.Front(); elt != nil; elt = elt.Next() {
		if elt.Value.(*Contact).id.Equals(contact.id) {
			break
		} else if elt.Value.(*Contact).id.Equals(table.node.id) {
			return
		}
	}
	if elt == nil {
		if bucket.Len() <= BucketSize {
			bucket.PushFront(contact)
		} else {
			// ping last seen node and handle for alive/dead
			last := bucket.Back().Value.(*Contact)
			if err := k.sendPingQuery(last); err == nil {
				/* TODO: Add new element to replacement cache list */
			} else {
				// Replace dead node with new live one
				bucket.Remove(bucket.Back())
				bucket.PushFront(contact)
			}
		}
	} else {
		bucket.MoveToFront(elt)
	}
}

func (k *kademlia) Serve() (err error) {
	rpc.Register(&kademliaCore{k})

	rpc.HandleHTTP()
	if l, err := net.Listen("tcp", k.routes.node.address); err == nil {
		go http.Serve(l, nil)
	}
	return
}

func (k *kademlia) call(contact *Contact, method string, args, reply interface{}) (err error) {
	if client, err := rpc.DialHTTP("tcp", contact.address); err == nil {
		err = client.Call(method, args, reply)
		if err == nil {
			k.Update(contact, k.routes)
		}
	}
	return
}

func (k *kademlia) sendPingQuery(node *Contact) (err error) {
	args := pingRequest{RPCHeader{&k.routes.node, k.NetworkID}}
	reply := pingResponse{}

	err = k.call(node, "kademliaCore.ping", &args, &reply)
	return
}

func (k *kademlia) sendFindNodeQuery(node *Contact, target NodeID, done chan []Contact) {
	args := findNodeRequest{RPCHeader{&k.routes.node, k.NetworkID}, target}
	reply := findNodeResponse{}

	if err := k.call(node, "kademliaCore.findNode", &args, &reply); err == nil {
		done <- reply.contacts
	} else {
		done <- []Contact{}
	}
}

func (k *kademlia) sendstoreQuery(node *Contact, domain string, typ string, ip net.IP) (err error) {
	args := storeRequest{RPCHeader{&k.routes.node, k.NetworkID}, domain, typ, ip}
	reply := storeResponse{}

	err = k.call(node, "kademliaCore.store", &args, &reply)
	return
}

func (k *kademlia) iterativeFindNode(target NodeID, delta int) (ret ContactRecList) {
	done := make(chan []Contact)

	// A heap of not-yet-queried *Contact structs
	frontier := &ContactHeap{}
	heap.Init(frontier)

	// A map of client values we've seen so far
	seen := make(map[string]bool)

	// Initialize the return list, frontier heap, and seen list with local nodes
	for _, node := range k.routes.findClosest(target, delta) {
		record := node
		ret = append(ret, record)
		heap.Push(frontier, *record.node)
		seen[record.node.id.String()] = true
	}

	// Start off delta queries
	pending := 0
	for i := 0; i < delta && frontier.Len() > 0; i++ {
		pending++
		node := frontier.Pop().(Contact)
		go k.sendFindNodeQuery(&node, target, done)
	}

	// Iteratively look for closer nodes
	for pending > 0 {
		nodes := <-done
		pending--
		for _, node := range nodes {
			// If we haven't seen the node before, add it
			if _, ok := seen[node.id.String()]; ok == false {
				ret = append(ret, &ContactRecord{&node, node.id.Xor(target)})
				heap.Push(frontier, node)
				seen[node.id.String()] = true
			}
		}

		for pending < delta && frontier.Len() > 0 {
			go k.sendFindNodeQuery(frontier.Pop().(*Contact), target, done)
			pending++
		}
	}

	sort.Sort(ret)
	if len(ret) > BucketSize {
		ret = ret[:BucketSize]
	}

	return
}

func (k *kademlia) iterativeStore(domain string, typ string, ip net.IP) {
	k.domains.storeRecord(domain, typ, ip) // store new/updated data locally
	target := NewNodeID(fmt.Sprintf("%x", domain))
	contacts := k.iterativeFindNode(target, 3)
	for _, contact := range contacts {
		if !contact.node.id.Equals(k.routes.node.id) {
			if err := k.sendstoreQuery(contact.node, domain, typ, ip); err != nil {
				log.Printf("Error sending store query for %s to %s\n", domain, contact.node)
			}
		}
	}
}

func (k *kademlia) HandleRPC(request, response *RPCHeader) error {
	if request.NetworkID != k.NetworkID {
		return errors.New(fmt.Sprintf("Expected network ID %s, got %s",
			k.NetworkID, request.NetworkID))
	}
	if request.Sender != nil {
		k.Update(request.Sender, k.routes)
	}
	response.Sender = &k.routes.node
	return nil
}

func (kc *kademliaCore) ping(args *pingRequest, response *pingResponse) (err error) {
	if err = kc.kad.HandleRPC(&args.RPCHeader, &response.RPCHeader); err == nil {
		log.Printf("ping from %s\n", args.RPCHeader)
	}
	return
}

func (kc *kademliaCore) store(args *storeRequest, response *storeResponse) (err error) {
	if err = kc.kad.HandleRPC(&args.RPCHeader, &response.RPCHeader); err == nil {
		kc.kad.domains.storeRecord(args.domain, args.typ, args.ip)
	}
	return
}

func (kc *kademliaCore) findNode(args *findNodeRequest, response *findNodeResponse) (err error) {
	if err = kc.kad.HandleRPC(&args.RPCHeader, &response.RPCHeader); err == nil {
		contacts := kc.kad.routes.findClosest(args.target, BucketSize)
		response.contacts = make([]Contact, contacts.Len())

		for i := 0; i < contacts.Len(); i++ {
			response.contacts[i] = *contacts[i].node
		}
	}
	return
}

func (kc *kademliaCore) findValue(args *findValueRequest, response *findValueResponse) (err error) {
	if err = kc.kad.HandleRPC(&args.RPCHeader, &response.RPCHeader); err == nil {
		val := kc.kad.domains.retrieve(args.domain, args.typ)
		if val != nil {
			response.ip = val
		} else {
			response.ip = nil
			target := NewNodeID(fmt.Sprintf("%x", args.domain))
			contacts := kc.kad.routes.findClosest(target, BucketSize)
			for i := 0; i < contacts.Len(); i++ {
				response.contacts[i] = *contacts[i].node
			}
		}
	}
	return
}
