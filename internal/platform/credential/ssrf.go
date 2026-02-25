package credential

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// ValidateRequestURL checks that the request URL is safe against SSRF attacks.
// Rules:
// - Scheme must be HTTPS
// - Host must match the credential's base URL host
// - No internal/private IP addresses
func ValidateRequestURL(requestURL, baseURL string) error {
	reqParsed, err := url.Parse(requestURL)
	if err != nil {
		return fmt.Errorf("invalid request URL: %w", err)
	}

	baseParsed, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	// HTTPS only
	if !strings.EqualFold(reqParsed.Scheme, "https") {
		return fmt.Errorf("only HTTPS URLs are allowed, got %q", reqParsed.Scheme)
	}

	// Host must match base URL
	reqHost := reqParsed.Hostname()
	baseHost := baseParsed.Hostname()
	if !strings.EqualFold(reqHost, baseHost) {
		return fmt.Errorf("request host %q does not match credential base URL host %q", reqHost, baseHost)
	}

	// Resolve host and check for internal IPs
	if err := checkNotInternalIP(reqHost); err != nil {
		return err
	}

	return nil
}

// checkNotInternalIP resolves the hostname and ensures it doesn't point to an internal IP.
func checkNotInternalIP(host string) error {
	ips, err := net.LookupHost(host)
	if err != nil {
		return fmt.Errorf("failed to resolve host %q: %w", host, err)
	}

	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if isInternalIP(ip) {
			return fmt.Errorf("host %q resolves to internal IP %s", host, ipStr)
		}
	}

	return nil
}

// isInternalIP returns true if the IP is in a private/reserved range.
func isInternalIP(ip net.IP) bool {
	privateRanges := []struct {
		network *net.IPNet
	}{
		{mustParseCIDR("127.0.0.0/8")},
		{mustParseCIDR("10.0.0.0/8")},
		{mustParseCIDR("172.16.0.0/12")},
		{mustParseCIDR("192.168.0.0/16")},
		{mustParseCIDR("169.254.0.0/16")},
		{mustParseCIDR("::1/128")},
		{mustParseCIDR("fc00::/7")},
		{mustParseCIDR("fe80::/10")},
	}

	for _, r := range privateRanges {
		if r.network.Contains(ip) {
			return true
		}
	}

	return false
}

func mustParseCIDR(cidr string) *net.IPNet {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(fmt.Sprintf("invalid CIDR: %s", cidr))
	}
	return network
}
