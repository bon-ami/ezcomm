package ezcomm

import (
	"context"
	"net"

	"github.com/pion/mdns"
	"golang.org/x/net/ipv4"
)

// MDNS_Server creats a server and waits for queries
// Parameters: chnErr can be nil, others must not be
func MDNS_Server(localName string, chnStp chan struct{}, chnErr chan error) {
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

	_, err = mdns.Server(ipv4.NewPacketConn(conn), &mdns.Config{
		LocalNames: []string{localName},
	})
	if err != nil {
		return
	}
	<-chnStp
}

// MDNS_Client creats a client and queries
// Parameters: channels can be nil
func MDNS_Client(localName string, chnAddr chan net.Addr,
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

	server, err := mdns.Server(ipv4.NewPacketConn(conn), &mdns.Config{})
	if err != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	if chnStp != nil {
		go func(chnStp chan struct{}) {
			select {
			case <-chnStp:
				cancel()
			case <-ctx.Done():
				break
			}
		}(chnStp)
	}
	_, src, err := server.Query(ctx, localName)
	chnAddr <- src
}
