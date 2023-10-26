package ezcomm

import (
	"net"
	"testing"
	"time"
)

func TestMdns(t *testing.T) {
	const localNameDef = "mdns.local"
	Init4Tests(t)
	var localName, remoteName string
	if len(*TstLcl) > 0 {
		localName = *TstLcl
	} else {
		localName = localNameDef
	}
	if len(*TstRmt) > 0 {
		remoteName = *TstRmt
	} else {
		remoteName = localNameDef
	}
	chnSvrErr := make(chan error, 1)
	chnSvrStp := make(chan struct{}, 1)
	go MdnsServer(localName, chnSvrStp, chnSvrErr)
	chnClntErr := make(chan error, 1)
	chnClntAddr := make(chan net.Addr, 1)
	chnClntStp := make(chan struct{}, 1)
	defer func() {
		close(chnSvrErr)
		close(chnSvrStp)
		close(chnClntErr)
		close(chnClntAddr)
		close(chnClntStp)
	}()
	go MdnsClient(remoteName, chnClntAddr, chnClntStp, chnClntErr)
	var timeout bool
	maxTries := 2 // wait for client and server
	for i := 0; i < maxTries; i++ {
		select {
		case <-time.After(TstTO):
			t.Error("timeout")
			if !timeout {
				timeout = true
				chnSvrStp <- struct{}{}
				chnClntStp <- struct{}{}
			}
			if maxTries == 3 {
				break
			} else {
				maxTries = 3
			}
		case err := <-chnSvrErr:
			if err != nil {
				t.Fatal("server failure.", err)
			}
			t.Log("server ends")
			if !timeout {
				timeout = true
				chnClntStp <- struct{}{}
			}
		case err := <-chnClntErr:
			if err != nil {
				t.Fatal("client failure.", err)
			} else {
				t.Log("client:", <-chnClntAddr)
			}
			t.Log("client ends")
			if !timeout {
				timeout = true
				chnSvrStp <- struct{}{}
			}
		}
	}
}
