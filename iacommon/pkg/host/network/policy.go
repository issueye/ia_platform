package network

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

var (
	ErrNetworkOperationNotSupported = errors.New("network operation is not supported")
	ErrNetworkSchemeNotAllowed      = errors.New("network scheme is not allowed")
	ErrNetworkHostNotAllowed        = errors.New("network host is not allowed")
	ErrNetworkPortNotAllowed        = errors.New("network port is not allowed")
	ErrNetworkRequestTooLarge       = errors.New("network request exceeds maximum allowed size")
	ErrInvalidNetworkRequest        = errors.New("invalid network request")
)

type Policy struct {
	Rights             []string
	AllowHosts         []string
	AllowPorts         []int
	AllowSchemes       []string
	AllowCIDRs         []string
	MaxConnections     int
	MaxInflightRequest int
	MaxBytesPerRequest int64
}

func (p Policy) ValidateHTTPRequest(req HTTPRequest) (*url.URL, error) {
	if strings.TrimSpace(req.URL) == "" {
		return nil, fmt.Errorf("%w: missing url", ErrInvalidNetworkRequest)
	}

	parsed, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("%w: parse url: %w", ErrInvalidNetworkRequest, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("%w: url must include scheme and host", ErrInvalidNetworkRequest)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if len(p.AllowSchemes) > 0 && !containsStringFold(p.AllowSchemes, scheme) {
		return nil, fmt.Errorf("%w: %s", ErrNetworkSchemeNotAllowed, scheme)
	}

	hostname := strings.ToLower(parsed.Hostname())
	if err := p.validateHost(hostname); err != nil {
		return nil, err
	}

	port, err := portForURL(parsed)
	if err != nil {
		return nil, err
	}
	if len(p.AllowPorts) > 0 && !containsInt(p.AllowPorts, port) {
		return nil, fmt.Errorf("%w: %d", ErrNetworkPortNotAllowed, port)
	}

	if p.MaxBytesPerRequest > 0 && int64(len(req.Body)) > p.MaxBytesPerRequest {
		return nil, fmt.Errorf("%w: %d > %d", ErrNetworkRequestTooLarge, len(req.Body), p.MaxBytesPerRequest)
	}

	return parsed, nil
}

func (p Policy) ValidateEndpoint(endpoint Endpoint) error {
	networkName := strings.ToLower(strings.TrimSpace(endpoint.Network))
	if networkName == "" {
		networkName = "tcp"
	}
	switch networkName {
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
	default:
		return fmt.Errorf("%w: unsupported network %s", ErrInvalidNetworkRequest, endpoint.Network)
	}

	host := strings.ToLower(strings.TrimSpace(endpoint.Host))
	if host == "" {
		return fmt.Errorf("%w: missing host", ErrInvalidNetworkRequest)
	}
	if err := p.validateHost(host); err != nil {
		return err
	}

	if endpoint.Port <= 0 || endpoint.Port > 65535 {
		return fmt.Errorf("%w: invalid port %d", ErrInvalidNetworkRequest, endpoint.Port)
	}
	if len(p.AllowPorts) > 0 && !containsInt(p.AllowPorts, endpoint.Port) {
		return fmt.Errorf("%w: %d", ErrNetworkPortNotAllowed, endpoint.Port)
	}

	return nil
}

func (p Policy) validateHost(hostname string) error {
	allowedByHost := len(p.AllowHosts) == 0 || containsStringFold(p.AllowHosts, hostname)
	if len(p.AllowCIDRs) == 0 {
		if !allowedByHost {
			return fmt.Errorf("%w: %s", ErrNetworkHostNotAllowed, hostname)
		}
		return nil
	}

	ip := net.ParseIP(hostname)
	allowedByCIDR := false
	if ip != nil {
		for _, cidr := range p.AllowCIDRs {
			_, network, err := net.ParseCIDR(cidr)
			if err != nil {
				continue
			}
			if network.Contains(ip) {
				allowedByCIDR = true
				break
			}
		}
	}

	if allowedByHost || allowedByCIDR {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrNetworkHostNotAllowed, hostname)
}

func portForURL(parsed *url.URL) (int, error) {
	if parsed == nil {
		return 0, fmt.Errorf("%w: nil url", ErrInvalidNetworkRequest)
	}
	if rawPort := parsed.Port(); rawPort != "" {
		port, err := strconv.Atoi(rawPort)
		if err != nil {
			return 0, fmt.Errorf("%w: invalid port %q", ErrInvalidNetworkRequest, rawPort)
		}
		return port, nil
	}

	switch strings.ToLower(parsed.Scheme) {
	case "http":
		return 80, nil
	case "https":
		return 443, nil
	default:
		return 0, fmt.Errorf("%w: unknown default port for scheme %s", ErrInvalidNetworkRequest, parsed.Scheme)
	}
}

func containsStringFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
