package kademlia

import (
	"container/list"
	"sort"
)

const BucketSize = 20

type RoutingTable struct {
	node    Contact
	buckets [IDLength * 8]*list.List
}

type ContactRecord struct {
	node    *Contact
	sortKey NodeID
}

func (rec *ContactRecord) Less(other interface{}) bool {
	return rec.sortKey.Less(other.(*ContactRecord).sortKey)
}

func NewRoutingTable(node *Contact) (ret *RoutingTable) {
	ret = new(RoutingTable)
	for i := 0; i < IDLength*8; i++ {
		ret.buckets[i] = list.New()
	}
	ret.node = *node
	return
}

func (table *RoutingTable) findClosest(target NodeID, count int) (ret ContactRecList) {

	bucket_num := target.Xor(table.node.id).PrefixLen()
	bucket := table.buckets[bucket_num]
	for elt := bucket.Front(); elt != nil; elt = elt.Next() {
		contact := elt.Value.(*Contact)
		ret = append(ret, &ContactRecord{contact, contact.id.Xor(target)})
	}

	for i := 1; (bucket_num-i >= 0 || bucket_num+i < IDLength*8) && ret.Len() < count; i++ {
		if bucket_num-i >= 0 {
			bucket = table.buckets[bucket_num-i]
			for elt := bucket.Front(); elt != nil; elt = elt.Next() {
				contact := elt.Value.(*Contact)
				ret = append(ret, &ContactRecord{contact, contact.id.Xor(target)})
			}
		}
		if bucket_num+i < IDLength*8 {
			bucket = table.buckets[bucket_num+i]
			for elt := bucket.Front(); elt != nil; elt = elt.Next() {
				contact := elt.Value.(*Contact)
				ret = append(ret, &ContactRecord{contact, contact.id.Xor(target)})
			}
		}
	}

	sort.Sort(ret)
	if ret.Len() > count {
		ret = ret[:count]
	}
	return
}
