// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package service

import "codeberg.org/miekg/dns"

// emptyRRs is a 2-element sequence that is empty. The type C is the logical
// container for each RR.
func emptyRRs[C any](yield func(C, dns.RR) bool) {}
