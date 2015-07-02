package kademlia

import (
  "net"
  "testing"
)

func TestStoreRecord(t *testing.T) {
  d := NewDomainStore()
  domain := "www.google.com"
  typ := "A"
  ip := net.ParseIP("74.125.224.72")
  d.StoreRecord(domain, typ, ip)

  if !d.data[domain][typ].Equal(ip) {
    t.Errorf("Data record %s does not match what was saved (%s)!", d.data[domain][typ].String(), ip.String())
  }
}
