package api

import (
	"os"
	"testing"
)

func TestIsTrustedProxy_DockerRanges(t *testing.T) {
	tests := []struct {
		ip     string
		wanted bool
	}{
		{"172.17.0.1", true},
		{"10.10.50.1", true},
		{"127.0.0.1", true},
		{"::1", true},
		{"1.2.3.4", false},
		{"192.168.1.1", false},
		{"not-an-ip", false},
	}
	for _, tt := range tests {
		if got := isTrustedProxy(tt.ip); got != tt.wanted {
			t.Errorf("isTrustedProxy(%q) = %v, want %v", tt.ip, got, tt.wanted)
		}
	}
}

func TestIsTrustedProxy_CustomCIDR(t *testing.T) {
	t.Setenv("TAGSHA_TRUSTED_PROXY_CIDRS", "192.168.1.0/24")
	_ = os.Getenv("TAGSHA_TRUSTED_PROXY_CIDRS")
	if isTrustedProxy("0.0.0.0") {
		t.Error("0.0.0.0 should never be trusted")
	}
}
