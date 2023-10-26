package ezcomm

import (
	"context"
	"net"
	"sync"

	"github.com/pion/mdns"
	"golang.org/x/net/ipv4"
)

// MdnsServer creats a server and waits for queries
// Parameters: chnErr can be nil, others must not
func MdnsServer(localName string, chnStp chan struct{}, chnErr chan error) {
	var err error
	defer func() {
		chnErr <- err
	}()
	addr, err := net.ResolveUDPAddr("udp", mdns.DefaultAddress)
	if err != nil {
		return
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return
	}
	defer conn.Close()

	_, err = mdns.Server(ipv4.NewPacketConn(conn), &mdns.Config{
		LocalNames: []string{localName},
	})
	if err != nil {
		return
	}
	<-chnStp
}

// MdnsClient creats a client and queries
// Parameters:
// localName must match server
// channels can be nil
func MdnsClient(localName string, chnAddr chan net.Addr,
	chnStp chan struct{}, chnErr chan error) {
	addr, err := net.ResolveUDPAddr("udp", mdns.DefaultAddress)
	defer func() {
		chnErr <- err
	}()
	if err != nil {
		return
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return
	}
	defer conn.Close()

	server, err := mdns.Server(ipv4.NewPacketConn(conn), &mdns.Config{})
	if err != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
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
	_, src, err := server.Query(ctx, localName)
	chnEnd <- struct{}{}
	chnAddr <- src
}
