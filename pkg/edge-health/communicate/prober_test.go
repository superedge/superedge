package communicate

import (
	"testing"
)

func TestTcpProbe(t *testing.T) {
	testcases := []struct {
		name     string
		ip       string
		port     int32
		timeout  int32
		expected bool
	}{
		{
			name:     "connect a unreachalbe ip port",
			ip:       "0.0.1.1",
			port:     128,
			timeout:  1,
			expected: false,
		},
		{
			name:     "connect a unresolveble domain",
			ip:       "test.my.test",
			port:     535,
			timeout:  1,
			expected: false,
		},
		{
			name:     "connect a normal ip port",
			ip:       "ifconfig.me",
			port:     80,
			timeout:  5,
			expected: true,
		},
	}
	for _, tc := range testcases {
		t.Log("run case", tc.name)

		r := tcpProbe(tc.ip, tc.port, tc.timeout)
		if tc.expected != r {
			t.Fatalf("test case %s failed, expect %v, actual %v", tc.name, tc.expected, r)
		}
	}
}

func TestIcmpProbe(t *testing.T) {
	testcases := []struct {
		name     string
		ip       string
		timeout  int32
		expected bool
	}{
		{
			name:     "connect a unreachalbe ip ",
			ip:       "0.0.1.1",
			timeout:  2,
			expected: false,
		},
		{
			name:     "connect a unresolveble domain",
			ip:       "test.my.test",
			timeout:  2,
			expected: false,
		},
		{
			name:     "connect a normal ip ",
			ip:       "ifconfig.me",
			timeout:  5,
			expected: true,
		},
	}
	for _, tc := range testcases {
		t.Log("run case", tc.name)

		r := icmpProbe(tc.ip, tc.timeout)
		if tc.expected != r {
			t.Skipf("test case %s failed, expect %v, actual %v", tc.name, tc.expected, r)
		}
	}
}
