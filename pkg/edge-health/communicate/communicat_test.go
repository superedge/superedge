package communicate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHandleProbe(t *testing.T) {
	testcases := []struct {
		name           string
		req            *Probe
		expectedResult []bool
	}{
		{
			name: "tcp connect a ip port",
			req: &Probe{Targets: []*Target{
				{
					IP:       "ifconfig.me",
					Port:     80,
					Protocol: "tcp",
				},
			}},
			expectedResult: []bool{true},
		},
		{
			name: "tcp connect multiple ip port",
			req: &Probe{Targets: []*Target{
				{
					IP:       "ifconfig.me",
					Port:     80,
					Protocol: "tcp",
				},
				{
					IP:       "127.0.0.1",
					Port:     60934,
					Protocol: "tcp",
				},
			}},
			expectedResult: []bool{true, false},
		},
		{
			name: "tcp connect single wrong ip port",
			req: &Probe{Targets: []*Target{
				{
					IP:       "127.0.0.1",
					Port:     60934,
					Protocol: "tcp",
				},
			}},
			expectedResult: []bool{false},
		},
		{
			name: "icmp single ip ",
			req: &Probe{Targets: []*Target{
				{
					IP:       "127.0.0.1",
					Protocol: "icmp",
				},
			}},
			expectedResult: []bool{true},
		},
		{
			name: "icmp multiple ip ",
			req: &Probe{Targets: []*Target{
				{
					IP:       "127.0.0.1",
					Protocol: "icmp",
				},
				{
					IP:       "99.99.99.99",
					Protocol: "icmp",
				},
			}},
			expectedResult: []bool{true, false},
		},
	}
	ce := &CommunicateEdge{
		10,
		2,
		1,
		51005,
		NewProberManager(2),
		&SourceInfo{
			"1.1.1.1",
			"testpod",
			"10.0.0.1",
			"testnode",
		},
	}
	s := httptest.NewServer(http.Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) { ce.HandleProbe(w, req) })))

	for _, tc := range testcases {
		t.Log("run case", tc.name)

		requestByte, _ := json.Marshal(tc.req)
		requestReader := bytes.NewReader(requestByte)
		req, err := http.NewRequest("POST", s.URL+"/result", requestReader)
		if err != nil {
			t.Fatalf("new request error, name %s", tc.name)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("http.DefaultClient.Do error, name %s", tc.name)
		}

		var probeResp ProbeResp

		err = json.NewDecoder(resp.Body).Decode(&probeResp)
		if err != nil {
			t.Fatalf("json.NewDecoder(resp.Body).Decode(&probeResp) error, name %s", tc.name)
		}
		resp.Body.Close()
		actualResult := []bool{}
		fmt.Printf("resp info: code %v targets length %v, source info %v\n", resp.StatusCode, len(probeResp.Targets), probeResp.SourceInfo)

		for _, target := range probeResp.Targets {
			fmt.Printf("target info ip %s port %d status %v\n", target.IP, target.Port, target.Normal)
			actualResult = append(actualResult, *target.Normal)
		}
		if !reflect.DeepEqual(tc.expectedResult, actualResult) {
			t.Skipf("test case %s failed, expect %v, actual %v", tc.name, tc.expectedResult, actualResult)
		}
	}
}
