package communicate

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"k8s.io/klog/v2"
)

type TCPProbeFunc func(ip string, port int32, timeout int32) bool
type ICMPProbeFunc func(ip string, timeout int32) bool

type ProberManager struct {
	TCPProbe     TCPProbeFunc
	ICMPProbe    ICMPProbeFunc
	ProbeTimeout int32
}

func NewProberManager(timeout int32) *ProberManager {
	return &ProberManager{
		TCPProbe:     tcpProbe,
		ICMPProbe:    icmpProbe,
		ProbeTimeout: timeout,
	}
}

func tcpProbe(ip string, port int32, timeout int32) bool {
	conn, err := net.DialTimeout(ProtoTCP, fmt.Sprintf("%s:%d", ip, port), time.Duration(timeout)*time.Second)
	defer func() {
		if conn != nil {
			if err := conn.Close(); err != nil {
				klog.ErrorS(err, "close tcp probe conn error")
			}
		}
	}()
	if err != nil {
		klog.ErrorS(err, "DialTimeout error")
	}
	return err == nil
}

// icmp use ipv4 send 3 packet and need receive 3 packet
func icmpProbe(ip string, timeout int32) bool {
	p, err := ping.NewPinger(ip)
	if err != nil {
		klog.ErrorS(err, "NewPinger(ip) error", "ip", ip)
	}
	p.SetPrivileged(true)
	p.SetNetwork("ip4")
	p.Interval = 300 * time.Millisecond
	p.Count = 3
	rchan := make(chan bool, 1)
	go func() {
		if err := p.Run(); err != nil {
			rchan <- false
			return
		}
		if p.Statistics().PacketsRecv == 3 {
			rchan <- true
			return
		}
		rchan <- false
	}()
	t := time.NewTimer(time.Duration(timeout) * time.Second)
	select {
	case <-t.C:
		klog.V(6).InfoS("icmp probe timeout", "ip", ip)
		return false
	case r := <-rchan:
		return r
	}
}

func (pm *ProberManager) Parallelize(workers, pieces int, doWorkPiece func(piece int)) {
	toProcess := make(chan int, pieces)
	for i := 0; i < pieces; i++ {
		toProcess <- i
	}
	close(toProcess)

	if pieces < workers {
		workers = pieces
	}

	wg := sync.WaitGroup{}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for piece := range toProcess {
				doWorkPiece(piece)
			}
		}()
	}
	wg.Wait()
}
