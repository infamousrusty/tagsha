// SPDX-License-Identifier: LGPL-3.0-or-later
// Copyright (C) 2026 infamousrusty

package api

import (
	"net"
	"os"
	"strings"
)

// trustedProxyCIDRs defines subnets from which X-Real-IP is accepted.
// Docker bridge networks use 172.16.0.0/12 and 10.0.0.0/8 by default.
// Override via TAGSHA_TRUSTED_PROXY_CIDRS (comma-separated) env var.
var trustedProxyCIDRs []*net.IPNet

func init() {
	raw := os.Getenv("TAGSHA_TRUSTED_PROXY_CIDRS")
	if raw == "" {
		raw = "172.16.0.0/12,10.0.0.0/8,127.0.0.1/32,::1/128"
	}
	for _, cidr := range strings.Split(raw, ",") {
		cidr = strings.TrimSpace(cidr)
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			trustedProxyCIDRs = append(trustedProxyCIDRs, network)
		}
	}
}

// isTrustedProxy returns true if the given IP address falls within a
// trusted proxy CIDR range.
func isTrustedProxy(remoteIP string) bool {
	ip := net.ParseIP(remoteIP)
	if ip == nil {
		return false
	}
	for _, network := range trustedProxyCIDRs {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}
