package ezcomm

import (
	"net"
	"testing"
)

func TestMdns(t *testing.T) {
	const localNameDef = "mdns.local"
	Init4Tests(t)
	var localName string
	if len(*TstLcl) > 0 {
		localName = *TstLcl
	} else {
		localName = localNameDef
	}
	chnSvrErr := make(chan error, 1)
	chnSvrStp := make(chan struct{}, 1)
	go MdnsServer(localName, chnSvrStp, chnSvrErr)
	chnClntErr := make(chan error, 1)
	chnClntAddr := make(chan net.Addr, 1)
	defer func() {
		close(chnSvrErr)
		close(chnSvrStp)
		close(chnClntErr)
		close(chnClntAddr)
	}()
	go MdnsClient(localName, chnClntAddr, nil, chnClntErr)
	for i := 0; i < 2; i++ { // wait for client and server
		select {
		case err := <-chnSvrErr:
			if err != nil {
				t.Fatal("server failure.", err)
			}
			t.Log("server ends")
		case err := <-chnClntErr:
			if err != nil {
				t.Fatal("client failure.", err)
			} else {
				t.Log("client:", <-chnClntAddr)
				chnSvrStp <- struct{}{}
			}
			t.Log("client ends")
		}
	}
}
