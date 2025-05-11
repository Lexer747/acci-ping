// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package ping

func (d Dropped) String() string {
	switch d {
	case BadResponse:
		return "Bad Response"
	case Timeout:
		return "Timeout"
	case DNSFailure:
		return "DNS Query Failed"
	case TestDrop:
		return "Testing A Dropped Packet :)"

	case NotDropped:
		fallthrough
	default:
		return ""
	}
}

func (at addressType) String() string {
	switch at {
	case _IP4:
		return "IP4"
	case _IP6:
		return "IP6"
	case _UDP4:
		return "UDP4"
	case _UDP6:
		return "UDP6"
	case _UNRESOLVED:
		return "unresolved"
	default:
		return "unknown"
	}
}
