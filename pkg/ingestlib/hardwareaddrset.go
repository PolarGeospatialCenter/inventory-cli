package ingestlib

import "net"

type HardwareAddrSet map[string]net.HardwareAddr

func NewHardwareAddrSet(macs ...net.HardwareAddr) HardwareAddrSet {
	set := make(HardwareAddrSet)
	set.Add(macs...)
	return set
}

func (s HardwareAddrSet) Add(macs ...net.HardwareAddr) {
	for _, mac := range macs {
		s[mac.String()] = mac
	}
}

func (s HardwareAddrSet) Get() []net.HardwareAddr {
	r := make([]net.HardwareAddr, 0, len(s))
	for _, mac := range s {
		r = append(r, mac)
	}
	return r
}

func (s HardwareAddrSet) Union(otherSet HardwareAddrSet) HardwareAddrSet {
	unionSet := make(HardwareAddrSet)
	unionSet.Add(s.Get()...)
	unionSet.Add(otherSet.Get()...)
	return unionSet
}
