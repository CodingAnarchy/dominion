package kademlia

import "net"

// DomainStore type contains a mapping of domain records to IP addresses.
type DomainStore struct {
	data map[string]map[string]net.IP
}

// NewDomainStore creates a new DomainStore type for storing domain record mapping.
func NewDomainStore() (ret *DomainStore) {
	ret = new(DomainStore)
	ret.data = make(map[string]map[string]net.IP)
	return
}

func (d *DomainStore) storeRecord(domain string, typ string, ip net.IP) {
	if d.data[domain] == nil {
		d.data[domain] = make(map[string]net.IP)
	}
	d.data[domain][typ] = ip
}

func (d *DomainStore) retrieve(domain string, typ string) (ip net.IP) {
	if d.data[domain] == nil || d.data[domain][typ] == nil {
		ip = nil
	} else {
		ip = d.data[domain][typ]
	}
	return
}
