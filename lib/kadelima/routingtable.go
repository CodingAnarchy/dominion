package kademlia

import (
  "container/list"
  "exp/iterable"
  "sort"
)

const BucketSize = 20

type RoutingTable struct {
  node Contact
  buckets [IDLength*8]*list.list
}

type ContactRecord struct {
  node *Contact
  sortKey nodeID
}

func (rec *ContactRecord) Less(other interface{}) bool {
  return rec.sortKey.Less(other.(*ContactRecord).sortKey)
}

func NewRoutingTable(node *Contact) (ret *RoutingTable) {
  ret = new(RoutingTable)
  for i:= 0; i < IDLength * 8; i++ {
    ret.buckets[i] = list.New()
  }
  ret.node = *node
  return
}

func (table *RoutingTable) Update(contact *Contact) {
  prefix_length := contact.id.Xor(table.node.id).PrefixLen()
  bucket := table.buckets[prefix_length]
  element := iterable.Find(bucket, func(x interface{}) bool {
    return x.(*Contact).id.Equals(table.node.id)
  })
  if element == nil {
    if bucket.Len() <= BucketSize {
      bucket.PushFront(contact)
    }
    // TODO: Handle insertion when the list is full by evicting old elements
    // if they don't respond to a ping
  } else {
    bucket.MoveToFront(element.(*list.Element))
  }
}

func copyToSlice(start, end *list.Element, slc []ContactRecord, target NodeID) {
  for elt := start; elt != end; elt = elt.Next() {
    contact := elt.Value.(*Contact)
    slc = append(slc, &ContactRecord{contact, contact.id.Xor(target)})
  }
}

func (table *RoutingTable) FindClosest(target NodeID, count int) (ret []ContactRecord) {

  bucket_num := target.Xor(table.node.id).PrefixLen()
  bucket := table.buckets[bucket_num]
  copyToSlice(bucket.Front(), nil, ret, target)

  for i:= 1; (bucket_num-i >= 0 || bucket_num+i < IDLength * 8) && ret.Len() < count; i++ {
    if bucket_num - i >= 0 {
      bucket = table.buckets[bucket_num - i]
      copyToSlice(bucket.Front(), nil, ret, target)
    }
    if bucket_num + i < IDLength * 8 {
      bucket = table.buckets[bucket_num + i]
      copyToSlice(bucket.Front(), nil, ret, target)
    }
  }

  sort.Sort(ret)
  if ret.Len() > count {
    ret = ret[:count]
  }
  return
}
