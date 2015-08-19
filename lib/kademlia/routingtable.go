package kademlia

import (
	"container/list"
	"sort"
)

const bucketSize = 20

// RoutingTable - store routing table in bucket lists
type RoutingTable struct {
	node    Contact
	buckets [idLength * 8]*list.List
}

// ContactRecord type is an individual contact record with node id for sortKey
type ContactRecord struct {
	node    *Contact
	sortKey NodeID
}

// Less - determine if a contact record is less than another off sortKey
func (rec *ContactRecord) Less(other interface{}) bool {
	return rec.sortKey.Less(other.(*ContactRecord).sortKey)
}

// NewRoutingTable - create new routing table for node
func NewRoutingTable(node *Contact) (ret *RoutingTable) {
	ret = new(RoutingTable)
	for i := 0; i < idLength*8; i++ {
		ret.buckets[i] = list.New()
	}
	ret.node = *node
	return
}

func (table *RoutingTable) findClosest(target NodeID, count int) (ret contactRecList) {

	bucketNum := target.Xor(table.node.id).PrefixLen()
	bucket := table.buckets[bucketNum]
	for elt := bucket.Front(); elt != nil; elt = elt.Next() {
		contact := elt.Value.(*Contact)
		ret = append(ret, &ContactRecord{contact, contact.id.Xor(target)})
	}

	for i := 1; (bucketNum-i >= 0 || bucketNum+i < idLength*8) && ret.Len() < count; i++ {
		if bucketNum-i >= 0 {
			bucket = table.buckets[bucketNum-i]
			for elt := bucket.Front(); elt != nil; elt = elt.Next() {
				contact := elt.Value.(*Contact)
				ret = append(ret, &ContactRecord{contact, contact.id.Xor(target)})
			}
		}
		if bucketNum+i < idLength*8 {
			bucket = table.buckets[bucketNum+i]
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
