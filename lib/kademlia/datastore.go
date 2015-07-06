package kademlia

import "net"

type DomainStore struct {
  data map[string]map[string]net.IP
}

func NewDomainStore() (ret *DomainStore) {
  ret = new(DomainStore)
  ret.data = make(map[string]map[string]net.IP)
  return
}

func (d *DomainStore) StoreRecord(domain string, typ string, ip net.IP) {
  if d.data[domain] == nil {
    d.data[domain] = make(map[string]net.IP)
  }
  d.data[domain][typ] = ip
}

func (d *DomainStore) Retrieve(domain string, typ string) (ip net.IP) {
  if d.data[domain] == nil || d.data[domain][typ] == nil {
    ip = nil
  } else {
    ip = d.data[domain][typ]
  }
  return
}
