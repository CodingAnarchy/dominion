package kademlia

import (
  "container/heap"
  "container/list"
  "fmt"
  "log"
  "net/http"
  "net"
  "net/rpc"
  "sort"
  "errors"
)

// Core Kademlia structs

type Kademlia struct {
  routes *RoutingTable
  NetworkID string
  domainStore map[string]map[string]net.IP
}

type KademliaCore struct {
  kad *Kademlia
}

// RPC Request and Response structs

type RPCHeader struct {
  Sender *Contact
  NetworkID string
}

type PingRequest struct {
  RPCHeader
}

type PingResponse struct {
  RPCHeader
}

type FindNodeRequest struct {
  RPCHeader
  target NodeID
}

type FindNodeResponse struct {
  RPCHeader
  contacts []Contact
}

type StoreRequest struct {
  RPCHeader
  domain string
  typ string
  ip net.IP
}

type StoreResponse struct {
  RPCHeader
}

// Data structures for internal use

// ContactHeap
type ContactHeap []Contact

func (c ContactHeap) Len() int           { return len(c) }
func (c ContactHeap) Less(i, j int) bool { return c[i].Less(c[j]) }
func (c ContactHeap) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

func (c *ContactHeap) Push(x interface{}) {
  *c =append(*c, x.(Contact))
}

func (c *ContactHeap) Pop() interface{} {
  old := *c
  n := len(old)
  x := old[n-1]
  *c = old[0:n-1]
  return x
}

// ContactRecList
type ContactRecList []*ContactRecord

func (cr ContactRecList) Len() int            { return len(cr) }
func (cr ContactRecList) Less(i, j int) bool  { return cr[i].Less(cr[j]) }
func (cr ContactRecList) Swap(i, j int)       { cr[i], cr[j] = cr[j], cr[i] }

// Kademlia functionality

func NewKademlia(self *Contact, networkID string) (ret *Kademlia) {
  ret = new(Kademlia)
  ret.routes = NewRoutingTable(self)
  ret.NetworkID = networkID
  ret.domainStore = make(map[string]map[string]net.IP)
  return
}

func (k * Kademlia) Update(contact *Contact, table *RoutingTable) {
  prefix_length := contact.id.Xor(table.node.id).PrefixLen()
  bucket := table.buckets[prefix_length]
  var elt *list.Element
  for elt = bucket.Front(); elt != nil; elt = elt.Next() {
    if elt.Value.(*Contact).id.Equals(table.node.id) {
      break
    }
  }
  if elt == nil {
    if bucket.Len() <= BucketSize {
      bucket.PushFront(contact)
    } else {
      // ping last seen node and handle for alive/dead
      last := bucket.Back().Value.(*Contact)
      if err:= k.sendPingQuery(last); err == nil {
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

func (k* Kademlia) Serve() (err error) {
  rpc.Register(&KademliaCore{k})

  rpc.HandleHTTP()
  if l, err := net.Listen("tcp", k.routes.node.address); err == nil {
    go http.Serve(l, nil)
  }
  return
}

func (k *Kademlia) Call(contact *Contact, method string, args, reply interface{}) (err error) {
  if client, err := rpc.DialHTTP("tcp", contact.address); err == nil {
    err = client.Call(method, args, reply)
    if err == nil {
      k.Update(contact, k.routes)
    }
  }
  return
}

func (k *Kademlia) sendPingQuery(node *Contact) (err error) {
  args := PingRequest{RPCHeader{&k.routes.node, k.NetworkID}}
  reply := PingResponse{}

  err = k.Call(node, "KademliaCore.Ping", &args, &reply)
  return
}

func (k *Kademlia) sendFindNodeQuery(node *Contact, target NodeID, done chan []Contact) {
  args := FindNodeRequest{RPCHeader{&k.routes.node, k.NetworkID}, target}
  reply := FindNodeResponse{}

  if err := k.Call(node, "KademliaCore.FindNode", &args, &reply); err == nil {
    done <- reply.contacts
  } else {
    done <- []Contact{}
  }
}

func (k *Kademlia) IterativeFindNode(target NodeID, delta int) (ret ContactRecList) {
  done := make(chan []Contact)

  // A heap of not-yet-queried *Contact structs
  frontier := &ContactHeap{}
  heap.Init(frontier)

  // A map of client values we've seen so far
  seen := make(map[string]bool)

  // Initialize the return list, frontier heap, and seen list with local nodes
  for _, node := range k.routes.FindClosest(target, delta) {
    record := node
    ret = append(ret, record)
    heap.Push(frontier, record.node)
    seen[record.node.id.String()] = true
  }

  // Start off delta queries
  pending := 0
  for i := 0; i < delta && frontier.Len() > 0; i++ {
    pending++
    go k.sendFindNodeQuery(frontier.Pop().(*Contact), target, done)
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

func (k *Kademlia) HandleRPC(request, response *RPCHeader) error {
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

func (kc *KademliaCore) Ping(args *PingRequest, response *PingResponse) (err error) {
  if err = kc.kad.HandleRPC(&args.RPCHeader, &response.RPCHeader); err == nil {
    log.Printf("Ping from %s\n", args.RPCHeader)
  }
  return
}

func (kc *KademliaCore) FindNode(args *FindNodeRequest, response *FindNodeResponse) (err error) {
  if err = kc.kad.HandleRPC(&args.RPCHeader, &response.RPCHeader); err == nil {
    contacts := kc.kad.routes.FindClosest(args.target, BucketSize)
    response.contacts = make([]Contact, contacts.Len())

    for i := 0; i < contacts.Len(); i++ {
      response.contacts[i] = *contacts[i].node
    }
  }
  return
}

func (kc *KademliaCore) Store(args *StoreRequest, response *StoreResponse) (err error) {
  if err = kc.kad.HandleRPC(&args.RPCHeader, &response.RPCHeader); err == nil {
    kc.kad.domainStore[args.domain][args.typ] = args.ip
  }
  return
}
