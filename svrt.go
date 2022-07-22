package ezcomm

import (
	"net"

	"gitee.com/bon-ami/eztools/v4"
)

// SvrTcp handles all clients and uses requested address to match them with user
/* routines
              SvrTcp  |  ezcomm
              ======= | =======
listening (Connected) | conn1:* ListeningTcp ListenTcp ConnectedTcp:2
===================== | =============================================
    |--------------------------------|-----------|

        1 incoming connection                    |---------->
                <---------------[2]channel, ezcomm----------|
    <-channel[0]|channel[1]>

                incoming traffic
    <-----chnLstn---------|<-------channel[1], ezcomm-------|
ChnConn[1]

                outgoing traffic
    |------------------------------channel[0], ezcomm------>|
ChnConn[0]
    |----------break 1 connection-------------------------->^
ChnConn[0]

    <----------1 connection broken--------------------------^
ChnConn[1]
    |--break 1 connection ->^

    |----------stop listening--------^
ChnConn[0]

    listening stopped, sending chnSvr             ^

    ^ (listening stopped and no connections in existence)
ChnConn[1]
*/
type SvrTcp struct {
	// chnErr is from EZ Comm when server stops
	chnErr chan error
	// chnSvr is the channel SvrTcp--FlowChnEnd->server
	//chnSvr chan RoutCommStruc
	// chnLstn is the channel for,
	//	conn--FlowChnRcv->SvrTcp, or,
	//	connected--FlowChnLst->SvrTcp
	chnLstn chan RoutCommStruc
	lstnr   net.Listener
	// chnStp [0/1] when server/clients stopped
	chnStp [2]chan struct{}
	/*// ChnConn traffic, for all connections
	ChnConn [2]chan RoutCommStruc*/
	/*// peeMapIn maps EZ Comm channel to requested address
	peeMapIn map[chan RoutCommStruc]string*/

	// following must to be set by user
	// LogFunc is not routine safe, for logging
	LogFunc FuncLog
	// ActFunc is invoked in one routine, to user.
	//	including FlowChnRcv, FlowChnEnd, FlowChnSnd
	ActFunc func(RoutCommStruc)
	// ConnFunc is run in a routine, on a connection incoming
	// also run on listening, to tell user of the actual address
	ConnFunc func([4]string)
}

// listening is routine for channels from EZ Comm
func (s SvrTcp) listening() {
	chnErr := s.chnErr
	chnLstn := s.chnLstn
	chnStp := s.chnStp
	peerMpO := make(map[string]chan RoutCommStruc)
	defer func() {
		//chnStp <- struct{}{}
		s.LogFunc("listening routine exit")
	}()
	svrDone := false
	chkDone := func() (noClients, allDone bool) {
		if len(peerMpO) < 1 {
			return true, svrDone
		} else {
			return false, false
		}
	}
	for {
		select {
		case /*err :=*/ <-chnErr:
			//s.LogFunc("chn err")
			chnStp[0] <- struct{}{}
			svrDone = true
			if _, d := chkDone(); d {
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
		case comm := <-chnLstn:
			//s.LogFunc("user requesting", comm)
			act1Conn := func(addr string) { //bool {
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
				peerMpO[comm.ReqAddr] = comm.Resp
				// from user
			case FlowChnSnd:
				act1Conn(comm.ReqAddr)
			case FlowChnEnd:
				if len(comm.ReqAddr) > 0 {
					act1Conn(comm.ReqAddr)
					delete(peerMpO, comm.ReqAddr)
				} else {
					for p, c := range peerMpO {
						c <- comm
						delete(peerMpO, p)
					}
				}
				// from EZ Comm
			case FlowChnRcv:
				s.ActFunc(comm)
			case FlowChnSnt:
				comm.Act = FlowChnSnd
				s.ActFunc(comm)
			case FlowChnDie:
				comm.Act = FlowChnEnd
				s.ActFunc(comm)
				noClients, allDone := chkDone()
				if noClients {
					chnStp[1] <- struct{}{}
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

// conn1 is routine for 1 connectin from EZ Comm
func (s SvrTcp) conn1(addr string, chn chan RoutCommStruc) {
	if s.chnLstn == nil {
		return
	}
	chnLstn := s.chnLstn
	for {
		comm := <-chn
		//s.LogFunc("connection got", addr, comm)
		comm.ReqAddr = addr
		switch comm.Act {
		case FlowChnSnd:
			comm.Act = FlowChnSnt
		case FlowChnEnd:
			comm.Act = FlowChnDie
		}
		chnLstn <- comm
		if comm.Act == FlowChnDie {
			s.LogFunc("connection dieing", addr, chnLstn, comm)
			break
		}
	}
}

// connected runs when a client comes in
func (s SvrTcp) connected(addr [4]string, chn [2]chan RoutCommStruc) {
	defer s.LogFunc("connection routine exit")
	if s.chnLstn == nil {
		s.LogFunc("NO listening channel!")
		return
	}
	addrReq := addr[1]
	s.chnLstn <- RoutCommStruc{
		Act:     FlowChnLst,
		ReqAddr: addrReq,
		Resp:    chn[0],
	}
	go s.ConnFunc(addr)
	s.conn1(addrReq, chn[1])
}

// Send is routine safe
//	act should be FlowChnSnd or FlowChnSndFil
func (s SvrTcp) Send(addr, data string, act int) {
	if s.chnLstn != nil {
		s.chnLstn <- RoutCommStruc{
			Act:     act,
			ReqAddr: addr,
			Data:    data,
		}
	}
}

// Listen returns whether successfully listening
// ConnFunc is called before returning, with only listening address as the first member of the slice.
func (s *SvrTcp) Listen(network, addr string) (err error) {
	if /*s.chnSvr != nil ||*/ s.chnStp[0] != nil || s.chnStp[1] != nil {
		return eztools.ErrIncomplete
	}
	s.chnErr = make(chan error, 1)
	s.chnLstn = make(chan RoutCommStruc, FlowComLen)
	for i := range s.chnStp {
		s.chnStp[i] = make(chan struct{}, 1)
	}
	s.lstnr, err = ListenTcp(s.LogFunc, s.connected,
		network, addr, ConnectedTcp, s.chnErr)
	if err != nil {
		return
	}
	s.ConnFunc([4]string{s.lstnr.Addr().String(), "", "", ""})
	//s.chnSvr = make(chan RoutCommStruc)
	//go ListeningTcp(s.LogFunc, s.chnSvr, lstnr)
	//s.peeMapIn = make(map[chan RoutCommStruc]string)
	/*for i := range s.ChnConn {
		s.ChnConn[i] = make(chan RoutCommStruc, FlowComLen)
	}*/
	go s.listening()
	return nil
}

// Wait returns when server/all clients stopped
//	user needs to run Disconnect() whatsever
func (s *SvrTcp) Wait(clients bool) {
	indx := 0
	switch clients {
	case true:
		indx = 1
	}
	<-s.chnStp[indx]
	s.chnStp[indx] = nil
	if !clients {
		s.chnErr = nil
	}
}

// Disconnect disconnects a/all connection(-s)
//	routine safe
func (s *SvrTcp) Disconnect(addr string) {
	if s.chnLstn != nil {
		s.chnLstn <- RoutCommStruc{
			Act:     FlowChnEnd,
			ReqAddr: addr,
		}
	}
}

// Stop stops listening
//	routine safe, except with Listen()
func (s *SvrTcp) Stop() {
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
