package kademlia

import (
  "container/heap"
  "container/vector"
  "fmt"
  "log"
  "http"
  "net"
  "os"
  "rpc"
  "sort"
)

type Kademlia struct {
  routes *RoutingTable
  NetworkID string
}

func NewKademlia(self *Contact, networkID string) (ret *Kademlia) {
  ret = new(Kademlia)
  ret.routes = NewRoutingTable(self)
  ret.NetworkID = networkID
  return
}

func (k* Kademlia) Serve() (err os.Error) {
  rpc.Register(&KademliaCore{k})

  rpc.HandleHTTP()
  if 1, err := net.Listen("tcp", k.routes.node.address); err == nil {
    go http.Serve(1, nil)
  }
  return
}

func (k *Kademlia) Call(contact *Contact, method string, args, reply interface{}) (err os.Error) {
  if client, err := rpc.DialHTTP("tcp", contact.address); err == nil {
    err = client.Call(method, args, reply)
    if err == nil {
      k.routes.Update(contact)
    }
  }
  return
}

func (k *Kademlia) sendQuery(node *Contact, target NodeID, done chan []Contact) {
  args := FindNodeRequest{RPCHeader(&k.routes.node, k.NetworkID), target}
  reply := FindNodeResponse{}

  if err := k.Call(node, "KademliaCore.FindNode", &args, &reply); err == nil {
    done <- reply.contacts
  } else {
    done <- []Contact{}
  }
}

func (k *Kademlia) IterativeFindNode(target NodeID, delta int) (ret *vector.Vector) {
  done := make(chan []Contact)

  // A vector of *ContactRecord structs
  ret = new(vector.Vector).Resize(0, BucketSize)

  // A heap of not-yet-queried *Contact structs
  frontier := new(vector.Vector).Resize(0, BucketSize)

  // A map of client values we've seen so far
  seen := make(map[string]bool)

  // Initialize the return list, frontier heap, and seen list with local nodes
  for node := range k.routes.FindClosest(target, delta).Iter() {
    record := node.(*ContactRecord)
    ret.Push(record)
    heap.push(frontier, record.node)
    seen[record.node.id.String()] = true
  }

  // Start off delta queries
  pending := 0
  for i := 0; i < delta && frontier.Len() > 0; i++ {
    pending++
    go k.sendQuery(frontier.Pop().(*Contact), target, done)
  }

  // Iteratively look for closer nodes
  for pending > 0 {
    nodes := <-done
    pending--
    for _, node := range nodes {
      // If we haven't seen the node before, add it
      if _, ok := seen[node.id.String()]; ok == false {
        ret.Push(&ContactRecord{&node, node.id.Xor(target)})
        heap.Push(frontier, node)
        seen[node.id.String()] = true
      }
    }

    for pending < delta && frontier.Len() > 0 {
      go k.sendQuery(frontier.Pop().(*Contact), target, done)
      pending++
    }
  }

  sort.Sort(ret)
  if ret.Len() > BucketSize {
    ret.Cut(BucketSize, ret.Len())
  }

  return
}

type RPCHeader struct {
  Sender *Contact
  NetworkID string
}

func (k *Kadelima) HandleRPC(request, response *RPCHeader) os.Error {
  if request.NetworkID != k.NetworkID {
    return os.NewError(fmt.Sprintf("Expected network ID %s, got %s",
                                    k.NetworkID, request.NetworkID))
  }
  if request.Sender != nil {
    k.routes.Update(request.Sender)
  }
  response.Sender = &k.routes.node
  return nil
}

type KademliaCore struct {
  kad *Kademlia
}

type PingRequest struct {
  RPCHeader
}

type PingResponse struct {
  RPCHeader
}

func (kc *KademliaCore) Ping(args *PingRequest, response *PingResponse) (err os.Error) {
  if err = kc.kad.HandleRPC(&args.RPCHeader, &response.RPCHeader); err == nil {
    log.Stderr("Ping from %s\n", args.RPCHeader)
  }
  return
}

type FindNodeRequest struct {
  RPCHeader
  target NodeID
}

type FindNodeResponse struct {
  RPCHeader
  contacts []Contact
}

func (kc *KademliaCore) FindNode(args *FindNodeRequest, response *FindNodeResponse) (err os.Error) {
  if err = kc.kad.HandleRPC(&args.RPCHeader, &response.RPCHeader); err == nil {
    contacts := kc.kad.routes.FindClosest(args.target, BucketSize)
    response.contacts = make([]Contact, contacts.Len())

    for i := 0; i < contacts.Len(); i++ {
      response.contacts[i] = *contacts.At(i).(*ContactRecord).node
    }
  }
  return
}
