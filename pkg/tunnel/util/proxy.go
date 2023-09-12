package util

import (
	"io"
	"net"
	"regexp"
	"strings"

	"k8s.io/klog/v2"
)

type Config struct {
	// ipMatchers represent all values in the NoProxy that are IP address
	// prefixes or an IP address in CIDR notation.
	ipMatchers []matcher

	// domainMatchers represent all values in the NoProxy that are a domain
	// name or hostname & domain name
	domainMatchers []matcher
}

// matcher represents the matching rule for a given value in the NO_PROXY list
type matcher interface {
	// match returns true if the host and optional port or ip and optional port
	// are allowed
	match(host, port string, ip net.IP) bool
}

type cidrMatch struct {
	cidr *net.IPNet
}

func (m cidrMatch) match(host, port string, ip net.IP) bool {
	return m.cidr.Contains(ip)
}

type ipMatch struct {
	ip   net.IP
	port string
}

func (m ipMatch) match(host, port string, ip net.IP) bool {
	if m.ip.Equal(ip) {
		return m.port == "" || m.port == port
	}
	return false
}

type domainMatch struct {
	host string
	port string

	matchHost bool
}

func (m domainMatch) match(host, port string, ip net.IP) bool {
	if strings.HasSuffix(host, m.host) || (m.matchHost && host == m.host[1:]) || strings.HasSuffix(host, m.host) {
		return m.port == "" || m.port == port
	}
	flag, err := regexp.MatchString(regexp.MustCompile(m.host[1:]).String(), host)
	if err != nil {
		return false
	}
	if flag {
		return m.port == "" || m.port == port
	}
	return false
}

// allMatch matches on all possible inputs
type allMatch struct{}

func (a allMatch) match(host, port string, ip net.IP) bool {
	return true
}

func NewHttpProxyConfig(proxy string) *Config {

	c := &Config{
		ipMatchers:     []matcher{},
		domainMatchers: []matcher{},
	}

	for _, p := range strings.Split(proxy, ",") {
		p = strings.ToLower(strings.TrimSpace(p))
		if len(p) == 0 {
			continue
		}

		if p == "*" {
			c.ipMatchers = []matcher{allMatch{}}
			c.domainMatchers = []matcher{allMatch{}}
			return c
		}

		// IPv4/CIDR, IPv6/CIDR
		if _, pnet, err := net.ParseCIDR(p); err == nil {
			c.ipMatchers = append(c.ipMatchers, cidrMatch{cidr: pnet})
			continue
		}

		// IPv4:port, [IPv6]:port
		phost, pport, err := net.SplitHostPort(p)
		if err == nil {
			if len(phost) == 0 {
				// There is no host part, likely the entry is malformed; ignore.
				continue
			}
			if phost[0] == '[' && phost[len(phost)-1] == ']' {
				phost = phost[1 : len(phost)-1]
			}
		} else {
			phost = p
		}
		// IPv4, IPv6
		if pip := net.ParseIP(phost); pip != nil {
			c.ipMatchers = append(c.ipMatchers, ipMatch{ip: pip, port: pport})
			continue
		}

		if len(phost) == 0 {
			// There is no host part, likely the entry is malformed; ignore.
			continue
		}

		// domain.com or domain.com:80
		// foo.com matches bar.foo.com
		// .domain.com or .domain.com:port
		// *.domain.com or *.domain.com:port
		if strings.HasPrefix(phost, "*.") {
			phost = phost[1:]
		}
		matchHost := false
		if phost[0] != '.' {
			matchHost = true
			phost = "." + phost
		}
		c.domainMatchers = append(c.domainMatchers, domainMatch{host: phost, port: pport, matchHost: matchHost})
	}
	return c
}

func (cfg *Config) UseProxy(addr string) bool {
	if len(addr) == 0 {
		return true
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return false
	}
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() {
			return false
		}
	}

	addr = strings.ToLower(strings.TrimSpace(host))

	if ip != nil {
		for _, m := range cfg.ipMatchers {
			if m.match(addr, port, ip) {
				return false
			}
		}
	}
	for _, m := range cfg.domainMatchers {
		if m.match(addr, port, ip) {
			return false
		}
	}
	return true
}

// ConnCopyAndClose process conn copy and close conn
func ConnCopyAndClose(dst, src net.Conn, uuid string) error {
	// copy data to remoteConn
	go func() {
		_, writeErr := io.Copy(dst, src)
		if writeErr != nil {
			klog.ErrorS(writeErr, "failed to copy data to remoteConn", STREAM_TRACE_ID, uuid)
		}
		dst.Close()
	}()

	// read data from remoteConn
	_, err := io.Copy(src, dst)
	if err != nil {
		klog.ErrorS(err, "failed to read data from remoteConn", STREAM_TRACE_ID, uuid)
		return err
	}
	src.Close()

	return nil
}
