package ezcomm

import (
	"net/netip"
	"testing"
	"time"
)

func TestMdns(t *testing.T) {
	const localNameDef = "mdns.local"
	Init4Tests(t)
	var localName, remoteName string
	if len(*TstLcl) > 0 && *TstLcl != Localhost+":" {
		localName = *TstLcl
	} else {
		localName = localNameDef
	}
	if len(*TstRmt) > 0 {
		remoteName = *TstRmt
	} else {
		remoteName = localNameDef
	}
	if *tstVerbose > 1 {
		t.Log("mdns server=", localName,
			";client=", remoteName,
			";timeout=", TstTO)
	}
	chnSvrErr := make(chan error, 1)
	chnSvrStp := make(chan struct{}, 1)
	go MdnsServer(localName, chnSvrStp, chnSvrErr)
	chnClntErr := make(chan error, 1)
	chnClntAddr := make(chan netip.Addr, 1)
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
				if *tstVerbose > 1 {
					t.Log("ending server")
				}
				chnSvrStp <- struct{}{}
				if *tstVerbose > 1 {
					t.Log("ending client")
				}
				chnClntStp <- struct{}{}
			}
			if maxTries == 3 {
				break
			} else {
				maxTries = 3
			}
		case errSvr := <-chnSvrErr:
			if errSvr != nil {
				t.Fatal("server failure.", errSvr)
			}
			t.Log("server ends")
			if !timeout {
				if *tstVerbose > 1 {
					t.Log("ending client")
				}
				timeout = true
				chnClntStp <- struct{}{}
			}
		case errClnt := <-chnClntErr:
			if errClnt != nil {
				t.Fatal("client failure.", errClnt)
			}
			t.Log("client ends from server:", <-chnClntAddr)
			if !timeout {
				if *tstVerbose > 1 {
					t.Log("ending server")
				}
				timeout = true
				chnSvrStp <- struct{}{}
			}
		}
	}
}
