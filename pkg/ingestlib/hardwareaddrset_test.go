package ingestlib

import (
	"net"
	"testing"
)

func TestHardwareAddrSet(t *testing.T) {
	m1, _ := net.ParseMAC("00:01:02:03:04:05")
	m2, _ := net.ParseMAC("01:01:02:03:04:05")
	m3, _ := net.ParseMAC("02:01:02:03:04:05")
	m4, _ := net.ParseMAC("03:01:02:03:04:05")
	s1 := NewHardwareAddrSet(m1, m2)
	s2 := NewHardwareAddrSet(m3, m4)

	if setLen := len(s1.Union(s2).Get()); setLen != 4 {
		t.Log(s1.Union(s2).Get(), setLen)
		t.Errorf("union doesn't return the expected results")
	}
}
