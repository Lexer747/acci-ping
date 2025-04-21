// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package ping

import "net"

const _IPv4len = 4
const _IPv6len = 16

func isIpv4(ip net.IP) bool {
	isZeros := func(p net.IP) bool {
		for i := range p {
			if p[i] != 0 {
				return false
			}
		}
		return true
	}
	if len(ip) == _IPv4len {
		return true
	}
	if len(ip) == _IPv6len &&
		isZeros(ip[0:10]) &&
		ip[10] == 0xff &&
		ip[11] == 0xff {
		return true
	}
	return false
}

func isIpv6(ip net.IP) bool {
	isZeros := func(p net.IP) bool {
		for i := range p {
			if p[i] != 0 {
				return false
			}
		}
		return true
	}
	if len(ip) == _IPv4len {
		return false
	}
	if len(ip) == _IPv6len &&
		isZeros(ip[0:10]) &&
		ip[10] == 0xff &&
		ip[11] == 0xff {
		return false
	}
	return true
}
