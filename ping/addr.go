// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package ping

import "net"

type addr struct {
	ip net.IP

	asUDP *net.UDPAddr
	asIP  *net.IPAddr
}

func New(addrType addressType, ip net.IP) *addr {
	// TODO Port?
	switch addrType {
	case _IP4, _IP6:
		return &addr{ip: ip, asIP: &net.IPAddr{IP: ip}}
	case _UDP4, _UDP6:
		return &addr{ip: ip, asUDP: &net.UDPAddr{IP: ip, Port: 0}}
	case _UNRESOLVED:
		// Will be resolved later.
		return &addr{ip: ip}
	default:
		panic("unknown addrType exhaustive:enforce")
	}
}

func (a *addr) Get() net.Addr {
	if a.asUDP != nil {
		return a.asUDP
	} else {
		return a.asIP
	}
}

func (a *addr) String() string {
	return a.ip.String()
}

func (a *addr) Equal(IP net.IP) bool {
	return a.ip.Equal(IP)
}
