package ezcomm

import (
	"net"
	"sync"

	"gitee.com/bon-ami/eztools/v6"
)

// SvrTCP handles all clients and uses requested address to match them with user
/* routines
              SvrTCP  |  ezcomm
              ======= | =======
listening (Connected) | connected:* ListenTcp ConnectedTcp:2
===================== | ===================================
    |-----------------------------------|

        1 incoming connection           |-------------->
                <-----[2]channel, ezcomm------------|
    <-channel[0]|channel[1]->

                incoming traffic
    <-----chnLstn-----------|<-channel[1], ezcomm---|
ChnConn[1]

                outgoing traffic
    |------------------channel[0], ezcomm---------->|
ChnConn[0]
    |----------break 1 connection------------------>^
ChnConn[0]

    <----------1 connection broken-----------------^
ChnConn[1]
    |--break 1 connection ->^

    |close listener
ChnConn[0]

      listening stopped, sending chnSvr^

    ^ (listening stopped and no connections in existence)
ChnConn[1]
*/
type SvrTCP struct {
	// chnErr is from EZ Comm when server stops
	chnErr chan error
	// chnLstn is the channel for,
	//	conn--FlowChnRcv->SvrTcp, or,
	//	connected--FlowChnLst->SvrTcp
	chnLstn  chan RoutCommStruc
	lockLstn sync.Mutex
	lstnr    net.Listener
	// chnStp [0/1] when server/clients stopped
	chnStp [2]chan struct{}

	// LogFunc is not routine safe, for logging
	// behavior undefined if not set
	LogFunc FuncLog
	// ActFunc is for FlowChnRcv, FlowChnEnd, FlowChnSnd to user.
	// behavior undefined if not set
	ActFunc func(RoutCommStruc)
	// ConnFunc runs upon an incoming connection
	//   It must not block
	// also runs on listening, to tell user of the actual address
	// behavior undefined if not set
	ConnFunc func([4]string)
}

// listening is routine for channels from EZ Comm
func (s *SvrTCP) listening() {
	peerMpO := make(map[string]chan RoutCommStruc)
	if eztools.Debugging && eztools.Verbose > 1 {
		s.LogFunc("entering TCP server listening routine")
		defer s.LogFunc("exiting TCP server listening routine")
	}
	defer func() {
		//s.chnStp <- struct{}{}
		s.lockLstn.Lock()
		close(s.chnLstn)
		s.chnLstn = nil
		s.lockLstn.Unlock()
	}()
	svrDone := false
	chkDone := func() (noClients, allDone bool) {
		//s.LogFunc("chkDone", peerMpO)
		if len(peerMpO) < 1 {
			return true, svrDone
		}
		return false, false
	}
	for {
		//s.LogFunc("looping")
		select {
		case /*err :=*/ <-s.chnErr:
			//s.LogFunc("chn err")
			s.chnStp[0] <- struct{}{}
			svrDone = true
			if noClients, allDone := chkDone(); allDone {
				if noClients && s.chnStp[1] != nil {
					s.chnStp[1] <- struct{}{}
				}
				return
			}
			//if err == eztools.ErrAbort {
			//return
			//}
			// TODO: do we need to handle it? how to check already End?
			/*s.ActFunc(RoutCommStruc{
				Act: FlowChnEnd,
				Err: err,
			})*/
		case comm := <-s.chnLstn:
			//s.LogFunc("user requesting", comm)
			act1Conn := func( /*addr*/ string) { //bool {
				chn, ok := peerMpO[comm.ReqAddr]
				if !ok {
					return
					/*comm.Err = eztools.ErrNoValidResults
					s.LogFunc("peers in rec", peerMpO)
					s.ActFunc(comm)*/
					//return true
				}
				chn <- comm
				//return false
			}
			switch comm.Act {
			// internally
			case FlowChnLst:
				/*if _, ok := peerMpO[comm.ReqAddr]; ok {
					s.LogFunc("Already connected!", comm.ReqAddr)
				} else {
					s.LogFunc("adding peer", comm.ReqAddr)
				}*/
				if s.chnStp[1] == nil {
					s.chnStp[1] = make(chan struct{}, 1)
				}
				peerMpO[comm.ReqAddr] = comm.Resp
				// from user
			case FlowChnSnd: //, FlowChnSndFil:
				act1Conn(comm.ReqAddr)
			case FlowChnEnd:
				if len(comm.ReqAddr) > 0 {
					act1Conn(comm.ReqAddr)
					if chn, ok := peerMpO[comm.ReqAddr]; ok {
						close(chn)
					}
					delete(peerMpO, comm.ReqAddr)
				} else {
					for p, c := range peerMpO {
						c <- comm
						close(c)
						delete(peerMpO, p)
					}
				}
				// from EZ Comm
			case FlowChnRcv:
				s.ActFunc(comm)
			case FlowChnSnt:
				comm.Act = FlowChnSnd
				s.ActFunc(comm)
			/*case FlowChnSntFil:
			comm.Act = FlowChnSndFil
			s.ActFunc(comm)*/
			case FlowChnDie:
				comm.Act = FlowChnEnd
				s.ActFunc(comm)
				noClients, allDone := chkDone()
				if noClients && s.chnStp[1] != nil {
					s.chnStp[1] <- struct{}{}
				}
				if allDone {
					return
				}
			default:
				s.LogFunc("UNKNOWN action", comm)
			}
		}
	}
}

// connected runs when a client comes in or blocked by anti-flood
func (s *SvrTCP) connected(addr [4]string, chn [2]chan RoutCommStruc) {
	defer close(chn[1])
	if s.chnLstn == nil {
		s.LogFunc("NO listening channel!")
		// should we panic?
		return
	}
	for _, c := range chn {
		if c == nil {
			s.LogFunc("probably flooding, no channel for", addr)
			close(chn[0])
			return
		}
	}
	// to record this address-channel pair
	s.chnLstn <- RoutCommStruc{
		Act:     FlowChnLst,
		ReqAddr: addr[1],
		Resp:    chn[0],
	}
	s.ConnFunc(addr)
	if eztools.Debugging && eztools.Verbose > 1 {
		s.LogFunc("entering TCP server connection routine")
		defer s.LogFunc("exiting TCP server connection routine")
	}
	for {
		comm := <-chn[1]
		//s.LogFunc("connection got", addr[1], comm)
		comm.ReqAddr = addr[1]
		switch comm.Act {
		case FlowChnSnd:
			comm.Act = FlowChnSnt
		/*case FlowChnSndFil:
		comm.Act = FlowChnSntFil*/
		case FlowChnEnd:
			comm.Act = FlowChnDie
		}
		s.chnLstn <- comm
		if comm.Act == FlowChnDie {
			s.LogFunc("connection ending",
				addr[1], s.chnLstn, comm)
			break
		}
	}
}

// Send is routine safe
//
//	act should be FlowChnSnd
func (s *SvrTCP) Send(addr string, data []byte) {
	s.lockLstn.Lock()
	if s.chnLstn != nil {
		s.chnLstn <- RoutCommStruc{
			Act:     FlowChnSnd,
			ReqAddr: addr,
			Data:    data,
		}
	}
	s.lockLstn.Unlock()
}

// HasStopped whether SvrTCP has Stop-ped
func (s *SvrTCP) HasStopped() bool {
	for _, ch := range s.chnStp {
		if ch != nil {
			return false
		}
	}
	return true
}

// Listen returns whether successfully listening
// ConnFunc is called before returning, with only listening address as the first member of the slice.
// ConnFunc may be called after a client incomes and is reported because of routine schedules.
func (s *SvrTCP) Listen(network, addr string) (err error) {
	//if [>s.chnSvr != nil ||<] s.chnStp[0] != nil || s.chnStp[1] != nil {
	if !s.HasStopped() {
		return eztools.ErrIncomplete
	}
	s.chnErr = make(chan error, 1)
	s.chnLstn = make(chan RoutCommStruc, FlowComLen)
	//for i := range s.chnStp {
	s.chnStp[0] = make(chan struct{}, 1)
	// [1] is to be created when connected
	//}
	s.lstnr, err = ListenTCP(s.LogFunc, s.connected,
		network, addr, Connected1Peer, s.chnErr)
	if err != nil {
		close(s.chnErr)
		s.chnErr = nil
		close(s.chnLstn)
		s.chnLstn = nil
		close(s.chnStp[0])
		s.chnStp[0] = nil
		return
	}
	s.ConnFunc([4]string{s.lstnr.Addr().String(), "", "", ""})
	//s.chnSvr = make(chan RoutCommStruc)
	go s.listening()
	return nil
}

// Wait returns when server/all clients stopped
//
//	user needs to run Disconnect() whatsever
func (s *SvrTCP) Wait(clients bool) {
	indx := 0
	switch clients {
	case true:
		indx = 1
	}
	if s.chnStp[indx] == nil {
		return
	}
	<-s.chnStp[indx]
	close(s.chnStp[indx])
	s.chnStp[indx] = nil
	if !clients {
		close(s.chnErr)
		s.chnErr = nil
	}
}

// Disconnect disconnects a/all connection(-s)
//
//	routine safe
func (s *SvrTCP) Disconnect(addr string) {
	s.lockLstn.Lock()
	if s.chnLstn != nil {
		s.chnLstn <- RoutCommStruc{
			Act:     FlowChnEnd,
			ReqAddr: addr,
		}
	}
	s.lockLstn.Unlock()
}

// Stop stops listening
//
//	routine safe, except with Listen()
func (s *SvrTCP) Stop() {
	if s.lstnr != nil {
		s.lstnr.Close()
		s.LogFunc("server stopped")
		/*s.chnSvr <- RoutCommStruc{
			Act: FlowChnEnd,
		}
		s.chnSvr = nil*/
	}
	/*if s.chnErr != nil {
		s.chnErr <- eztools.ErrAbort
	}*/
}
