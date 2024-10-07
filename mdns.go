package ezcomm

import (
	"context"
	"net"
	"net/netip"
	"sync"
	"time"

	"gitee.com/bon-ami/eztools/v6"
	"github.com/pion/mdns/v2"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// MdnsServer creats a server and waits for queries
// Parameters: chnErr can be nil, others must not
// Return values: not to be used, except for internal MdnsClient
// for internal usage, when localName is empty and chnStp is nil,
// server does not wait and is for MdnsClient.
func MdnsServer(localName string, chnStp chan struct{},
	chnErr chan error) (server *mdns.Conn, err error) {
	var l4, l6 *net.UDPConn
	defer func() {
		if eztools.Debugging && eztools.Verbose > 2 {
			eztools.Log("mdns server exiting", err)
		}
		if chnErr != nil {
			chnErr <- err
		}
		if chnStp != nil {
			if l4 != nil {
				l4.Close()
			}
			if l6 != nil {
				l6.Close()
			}
		}
	}()
	addr4, err := net.ResolveUDPAddr("udp4", mdns.DefaultAddressIPv4)
	if err != nil {
		return
	}

	addr6, err := net.ResolveUDPAddr("udp6", mdns.DefaultAddressIPv6)
	if err != nil {
		return
	}

	l4, err = net.ListenUDP("udp4", addr4)
	if err != nil {
		return
	}

	l6, err = net.ListenUDP("udp6", addr6)
	if err != nil {
		return
	}

	var cfg *mdns.Config
	if len(localName) < 1 {
		cfg = &mdns.Config{}
	} else {
		cfg = &mdns.Config{
			LocalNames: []string{localName},
		}
	}
	server, err = mdns.Server(ipv4.NewPacketConn(l4), ipv6.NewPacketConn(l6), cfg)
	if err != nil {
		return
	}
	if chnStp != nil {
		if eztools.Debugging && eztools.Verbose > 1 {
			eztools.Log("mdns server", localName, "running")
		}
		<-chnStp
	} else {
		if eztools.Debugging && eztools.Verbose > 1 {
			eztools.Log("mdns server for client running")
		}
	}
	return
}

// MdnsClient creats a client and queries
// Parameters:
// localName must match server
// channels can be nil
func MdnsClient(localName string, chnAddr chan netip.Addr,
	chnStp chan struct{}, chnErr chan error, timeout time.Duration) {
	var src netip.Addr
	var err error
	defer func() {
		if eztools.Debugging && eztools.Verbose > 2 {
			eztools.Log("mdns client exiting", err)
		}
		chnErr <- err
		chnAddr <- src
	}()
	server, err := MdnsServer("", nil, nil)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	var (
		cancelled bool
		lock      sync.Mutex
	)
	cancelDo := func() {
		lock.Lock()
		if !cancelled {
			cancelled = true
			cancel()
		}
		lock.Unlock()
	}
	defer cancelDo()
	chnEnd := make(chan struct{}, 1)
	defer close(chnEnd)
	if chnStp != nil {
		go func() {
			select {
			case <-chnStp:
				cancelDo()
			case <-chnEnd:
			}
		}()
	}
	if eztools.Debugging && eztools.Verbose > 1 {
		eztools.Log("mdns client", localName, "running")
	}
	_, src, err = server.QueryAddr(ctx, localName)
	chnEnd <- struct{}{}
}
