package probe

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

type Result struct {
	Target string
	IP     string
	Seq    int
	OK     bool
	RTT    time.Duration
	Err    error
}

type targetSock struct {
	ip   string
	isV4 bool
	addr *net.IPAddr
	conn *icmp.PacketConn
}

type ICMPProber struct {
	ID      int
	Timeout time.Duration

	mu    sync.Mutex
	socks map[string]*targetSock
}

func NewICMPProber(timeout time.Duration) *ICMPProber {
	return &ICMPProber{
		ID:      os.Getpid() & 0xffff,
		Timeout: timeout,
		socks:   make(map[string]*targetSock),
	}
}

func (p *ICMPProber) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var firstErr error
	for _, s := range p.socks {
		if s == nil || s.conn == nil {
			continue
		}
		if err := s.conn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	p.socks = make(map[string]*targetSock)
	return firstErr
}

func (p *ICMPProber) Probe(target string, seq int) Result {
	resolve := resolveTarget(target, seq)
	if resolve.Err != nil {
		return resolve
	}

	ip := net.ParseIP(resolve.IP)
	if ip == nil {
		resolve.Err = fmt.Errorf("invalid resolved ip: %q", resolve.IP)
		return resolve
	}
	isV4 := ip.To4() != nil

	sock, err := p.getSocket(target, resolve.IP, isV4, ip)
	if err != nil {
		resolve.Err = err
		return resolve
	}

	msg := icmp.Message{
		Type: echoRequestType(isV4),
		Code: 0,
		Body: &icmp.Echo{
			ID:   p.ID,
			Seq:  seq,
			Data: []byte("rttlog"),
		},
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		resolve.Err = err
		return resolve
	}

	start := time.Now()
	if _, err := sock.conn.WriteTo(msgBytes, sock.addr); err != nil {
		resolve.Err = err
		return resolve
	}

	_ = sock.conn.SetReadDeadline(time.Now().Add(p.Timeout))

	reply := make([]byte, 1500)
	for {
		n, _, err := sock.conn.ReadFrom(reply)
		if err != nil {
			resolve.Err = err
			return resolve
		}

		parsedMsg, err := icmp.ParseMessage(getProtocolNumber(isV4), reply[:n])
		if err != nil {
			continue
		}

		if parsedMsg.Type != echoReplyType(isV4) {
			continue
		}

		body, ok := parsedMsg.Body.(*icmp.Echo)
		if !ok {
			continue
		}

		if body.ID != p.ID || body.Seq != seq {
			continue
		}

		return Result{
			Target: target,
			IP:     resolve.IP,
			Seq:    seq,
			OK:     true,
			RTT:    time.Since(start),
		}
	}
}

func (p *ICMPProber) getSocket(target, resolvedIP string, isV4 bool, parsedIP net.IP) (*targetSock, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	s := p.socks[target]
	if s != nil && s.ip == resolvedIP && s.isV4 == isV4 && s.conn != nil {
		return s, nil
	}

	if s != nil && s.conn != nil {
		_ = s.conn.Close()
	}

	conn, err := openICMPSocket(isV4)
	if err != nil {
		return nil, err
	}

	ns := &targetSock{
		ip:   resolvedIP,
		isV4: isV4,
		addr: &net.IPAddr{IP: parsedIP},
		conn: conn,
	}
	p.socks[target] = ns
	return ns, nil
}

func resolveTarget(target string, seq int) Result {
	res := Result{
		Target: target,
		Seq:    seq,
	}

	ipaddr, err := net.ResolveIPAddr("ip", target)
	if err != nil {
		res.Err = err
		return res
	}

	res.IP = ipaddr.IP.String()
	return res
}

func openICMPSocket(isV4 bool) (*icmp.PacketConn, error) {
	network := "ip4:icmp"
	if !isV4 {
		network = "ip6:ipv6-icmp"
	}
	return icmp.ListenPacket(network, "")
}

func echoRequestType(isV4 bool) icmp.Type {
	if isV4 {
		return ipv4.ICMPTypeEcho
	}
	return ipv6.ICMPTypeEchoRequest
}

func echoReplyType(isV4 bool) icmp.Type {
	if isV4 {
		return ipv4.ICMPTypeEchoReply
	}
	return ipv6.ICMPTypeEchoReply
}

func getProtocolNumber(isV4 bool) int {
	if isV4 {
		return 1
	}
	return 58
}
