package ezcomm

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"gitee.com/bon-ami/eztools/v4"
)

// Uis is for ezcomm (main module) -> users (UI)
//type Uis interface {
// FuncLog is log function
type FuncLog func(...any)

// FuncConn is run when Connected.
// addr=address info.
//	[0]: parsed local
//	[1]: remote
//	[2]: requested protocol
//	[3]: requested address
type FuncConn func(addr [4]string, chn [2]chan RoutCommStruc)

//}

var (
	// AntiFlood is the limit of traffic from a peer
	AntiFlood struct {
		// Limit is for incoming traffic per second.
		// negative values means limitless.
		// 0=any incoming traffic causes disconnection
		Limit int64
		// Period is to disconnect the peer immediatelyi,
		// if it connects again since previous flood.
		// non-positive values means forever.
		Period int64
		// refused records time refused flood
		refused map[string]int64
		lock    sync.Mutex
	}
)

type RoutCommStruc struct {
	// ReqAddr is address from user
	ReqAddr string
	// Act=FlowChn*
	Act int
	// PeerUdp is filled, and required for FlowChnSnd by EZ Comm
	PeerUdp *net.UDPAddr
	// PeerTcp is filled, but not required by EZ Comm
	PeerTcp net.Addr
	// I/O raw message
	Data []byte
	// Resp for SvrTcp and flow only
	Resp chan RoutCommStruc
	Err  error
}

func replyWtErr(comm RoutCommStruc, err error, chn chan RoutCommStruc) {
	//logFunc("UDP", cmd.PeerUdp, n, err)
	comm.Err = err
	chn <- comm
	//logFunc("UDP sent", comm)
}

// ConnectedUdp works for UDP, when remote can change
//	It blocks.
// chn[1] -> ui: FlowChnRcv, FlowChnSnd, FlowChnEnd
// chn[0] <- ui: FlowChnSnd, FlowChnEnd
func ConnectedUdp(logFunc FuncLog, chn [2]chan RoutCommStruc, conn *net.UDPConn) {
	defer conn.Close()
	buf := make([]byte, FlowRcvLen)
	lcl := conn.LocalAddr().String()
	if eztools.Debugging && eztools.Verbose > 1 {
		logFunc("entering udp routine", lcl)
		defer func() {
			logFunc("exiting udp routine", lcl)
		}()
	}
	go func() {
		if eztools.Debugging && eztools.Verbose > 1 {
			logFunc("entering routine", lcl)
		}
		defer func() {
			chn[1] <- RoutCommStruc{
				Act: FlowChnEnd,
			}
			if eztools.Debugging && eztools.Verbose > 1 {
				logFunc("exiting routine", lcl)
			}
		}()
		floodRecs := make(map[string][2]int64)
		for {
			//logFunc("receiving UDP", conn.LocalAddr())
			n, addr, err := conn.ReadFromUDP(buf)
			//logFunc("received UDP", n, addr, err, buf)
			if err != nil &&
				(errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed)) {
				return
			}
			if addr == nil {
				// such as an err buf is smaller than data
				continue
			}
			comm := RoutCommStruc{
				Act: FlowChnRcv,
				Err: err,
			}
			peerAddr := addr.String()
			peerHost, _, err := net.SplitHostPort(peerAddr)
			if err != nil {
				logFunc(peerAddr, err)
			}
			if floodChk(peerHost) {
				continue
			}
			ok, recs := floodCtrl(func() [2]int64 {
				recs, ok := floodRecs[peerHost]
				//logFunc("checking flood", peerHost, ok, recs)
				if ok {
					return recs
				}
				return [2]int64{0, 0}
			}, addr.String())
			if ok {
				//logFunc("flooding", peerAddr)
				continue
			}
			floodRecs[peerHost] = recs
			if err == nil {
				comm.Data = buf[:n]
				comm.PeerUdp = addr
			}
			chn[1] <- comm
			/*if ui != nil {
				ui.Rcv(comm)
			}*/
		}
	}()
	//logFunc("listening on UDP", ChanComm)
	var err error
	for {
		//logFunc("command udp", chn, "waiting")
		cmd := <-chn[0]
		//logFunc("command udp", chn, cmd)
		switch cmd.Act {
		case FlowChnSnd:
			if eztools.Debugging && eztools.Verbose > 2 {
				logFunc("sending", cmd)
			}
			// TODO: confirm to anti-flood
			if cmd.PeerUdp == nil {
				err = eztools.ErrIncomplete
			} else {
				_, err = conn.WriteToUDP(cmd.Data, cmd.PeerUdp)
			}
			replyWtErr(cmd, err, chn[1])
		case FlowChnEnd:
			//logFunc("exiting", lcl)
			return
		}
	}
}

/*func ListeningTcp(logFunc FuncLog, chn [2]chan struct{},
lstnr net.Listener) {
	defer logFunc("exiting TCP server routine", lstnr.Addr().String())
	<-chn[0]
	lstnr.Close()
	return
}*/

// floodCtrl check for flood in a connection
// Parameters: getRecs[0]=previous start of flood check; [1]=number of packets within a second
// return values: flooding; updated data for getRecs
func floodCtrl(getRecs func() [2]int64,
	peer string) (bool, [2]int64) {
	if AntiFlood.Limit < 0 {
		return false, [2]int64{0, 0}
	}
	floodRecs := getRecs()
	curr := time.Now().Unix()
	if (curr - floodRecs[0]) > 0 {
		floodRecs[0] = curr
		floodRecs[1] = 0
		return false, floodRecs
	}
	if floodRecs[1] < AntiFlood.Limit {
		floodRecs[1]++
		return false, floodRecs
	}
	if AntiFlood.Period > 0 {
		AntiFlood.lock.Lock()
		if AntiFlood.refused == nil {
			AntiFlood.refused = make(map[string]int64)
		}
		AntiFlood.refused[peer] = curr
		AntiFlood.lock.Unlock()
	}
	return true, floodRecs
}

func floodChk(peerAddr string) bool {
	if AntiFlood.Limit < 0 {
		return false
	}
	AntiFlood.lock.Lock()
	if AntiFlood.refused == nil {
		AntiFlood.refused = make(map[string]int64)
	}
	prev, ok := AntiFlood.refused[peerAddr]
	AntiFlood.lock.Unlock()
	if !ok {
		return false
	}
	if AntiFlood.Period < 1 ||
		AntiFlood.Period+prev > time.Now().Unix() {
		return true
	}
	AntiFlood.lock.Lock()
	delete(AntiFlood.refused, peerAddr)
	AntiFlood.lock.Unlock()
	return false
}

// ConnectedTcp also works for UDP, if remote does not change
// Parameters:
//	logFunc for logging
//	connFunc is callback function upon entrance of this function.
//		It blocks the routine.
//		If flood detected, it is called with local and remote addresses, requested protocol and address, and nil channels
//	conn is the connection
//	addrReq is address user requested, and varies between local/remote,
//		when user creates a server (listen) or a client
// -> ui: connFunc(), FlowChnRcv, FlowChnSnd, FlowChnEnd
// <- ui: FlowChnSnd, FlowChnEnd // this may block routine from exiting, if too much incoming traffic not read
func ConnectedTcp(logFunc FuncLog, connFunc FuncConn, conn net.Conn, addrReq [2]string) {
	defer conn.Close()
	peer := conn.RemoteAddr()
	peerAddr := peer.String()
	peerHost, _, err := net.SplitHostPort(peerAddr)
	if err != nil {
		logFunc(peerAddr, err)
	}
	localAddr := conn.LocalAddr().String()
	var chn [2]chan RoutCommStruc
	connCB := func() {
		connFunc([...]string{localAddr, peerAddr,
			addrReq[0], addrReq[1]}, chn)
	}
	if floodChk(peerHost) {
		connCB()
		return
	}
	for i := range chn {
		chn[i] = make(chan RoutCommStruc, FlowComLen)
	}
	//if ui != nil {
	connCB()
	//}
	buf := make([]byte, FlowRcvLen)
	go func() {
		if eztools.Debugging && eztools.Verbose > 1 {
			logFunc("entering routine", localAddr)
		}
		var floodRecs [2]int64
		var comm RoutCommStruc
		defer func() {
			if eztools.Debugging && eztools.Verbose > 1 {
				logFunc("exiting TCP", localAddr, "routine peer", peerAddr)
			}
			comm.Act = FlowChnEnd
			chn[1] <- comm
		}()
		for {
			//logFunc("TCP", localAddr, "to read")
			n, err := conn.Read(buf)
			//logFunc("TCP", localAddr, "read from", peerAddr, n, err)
			comm = RoutCommStruc{
				Act:     FlowChnRcv,
				PeerTcp: peer,
				Err:     err,
			}
			floodChkTcp := func() bool {
				var ok bool
				ok, floodRecs = floodCtrl(
					func() [2]int64 {
						return floodRecs
					}, peerHost)
				if ok {
					//logFunc("flooding", peerHost)
					//if ui != nil {
					comm.Err = eztools.ErrAccess
					//ui.Ended(*comm)
					//}
					return true
				}
				return false
			}

			if err == nil {
				comm.Data = make([]byte, n)
				copy(comm.Data, buf[:n])
			} else {
				if errors.Is(err, io.EOF) ||
					errors.Is(err, net.ErrClosed) {
					//if ui != nil {
					/*comm.Act = FlowChnEnd
					chn[1] <- comm*/
					//ui.Ended(comm)
					//}
					break
				}
				// reading on a closed TCP connection also reaches here
				break
			}
			if floodChkTcp() {
				comm.Err = eztools.ErrAccess
				break
			}
			chn[1] <- comm
			/*if ui != nil {
				ui.Rcv(comm)
			}*/
		}
	}()
	for {
		cmd := <-chn[0]
		//logFunc(cmd, "got from user for", peerAddr)
		switch cmd.Act {
		case FlowChnSnd:
			cmd.PeerTcp = peer
			// TODO: confirm to anti-flood
			_, err := conn.Write(cmd.Data)
			replyWtErr(cmd, err, chn[1])
		case FlowChnEnd:
			logFunc("exiting TCP connection routine", localAddr, "peer", peerAddr)
			return
		}
	}
}

/*func log(ui Uis, sth ...any) {
	if !eztools.Debugging {
		return
	}
	if ui != nil {
		ui.Log(sth...)
	} else {
		eztools.LogPrint(sth...)
	}
}*/

// Client is mainly for TCP and Unix(not -gram). It ends immediately.
//   lstnFunc() needs to handle procedures afterwards.
//   returned conn needs to Close() by user.
// UDP, IP and Unixgram will be tricky to track peer information,
//   and preferrably use Listen*()
func Client(logFunc FuncLog, connFunc FuncConn,
	network, rmtAddr string,
	lstnFunc func(logFunc FuncLog, connFunc FuncConn,
		conn net.Conn, rmtAddr [2]string)) (conn net.Conn, err error) {
	conn, err = net.Dial(network, rmtAddr)
	if err != nil {
		return
	}
	//defer conn.Close()
	if lstnFunc != nil {
		go lstnFunc(logFunc, connFunc, conn, [2]string{network, rmtAddr})
	}
	return
}

// ListenIp listens to IP. It ends immediately.
// Parameters:
//	network is socket type, "ip", "ip4" or "ip6"
//	address is for local, [IP or DN][:protocol number or name]
func ListenIp(network, address string) (*net.IPConn, error) {
	var (
		addr *net.IPAddr
		err  error
	)
	if len(address) > 0 {
		addr, err = net.ResolveIPAddr(network, address)
		if err != nil {
			return nil, err
		}
	}
	return net.ListenIP(network, addr)
}

// ServerUnix listens to unixgram. It ends immediately.
// Parameters:
//	network is socket type, "unixgram"
//	address is for local, [IP or DN][:port number or name]
func ListenUnixgram(network, address string) (*net.UnixConn, error) {
	var (
		addr *net.UnixAddr
		err  error
	)
	if len(address) > 0 {
		addr, err = net.ResolveUnixAddr(network, address)
		if err != nil {
			return nil, err
		}
	}
	return net.ListenUnixgram(network, addr)
}

// ListenUdp listens to UDP. It ends immediately.
//   *net.UDPConn needs to handle procedures afterwards.
// Parameters:
//	network is socket type, "udp", "udp4" or "udp6"
//	address is for local, [IP or DN][:port number or name]
func ListenUdp(network, address string) (*net.UDPConn, error) {
	var (
		addr *net.UDPAddr
		err  error
	)
	if len(network) < 1 {
		network = "udp"
	}
	if len(address) > 0 {
		addr, err = net.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
	}
	return net.ListenUDP(network, addr)
}

// ListenTcp listens to TCP. It ends immediately.
//   Simply close the listener to stop it.
//   accepted() needs to handle procedures for incoming connections.
//     It blocks listening routine, so that errChan gets feedback
//     always before accepted().
//   connFunc is for accepted() only.
// Parameters:
//	network is socket type, "tcp", "tcp4", "tcp6",
//		"unix" or "unixpacket"
//	fun handles incoming connections for TCP/unix and all connections for UDP
//	errChan sends Accept errors
func ListenTcp(logFunc FuncLog, connFunc FuncConn,
	network, address string, accepted func(FuncLog,
		FuncConn, net.Conn, [2]string), errChan chan error) (net.Listener, error) {
	if accepted == nil {
		eztools.LogFatal("no function to handle server")
	}
	if len(network) < 1 {
		network = "tcp"
	}
	//log("to serve")
	lstnr, err := net.Listen(network, address)
	if err != nil {
		return lstnr, err
	}
	//log("serving", lstnr)
	//defer lstnr.Close()
	go func() {
		if eztools.Debugging && eztools.Verbose > 1 {
			logFunc("entering listener routine")
		}
		var (
			err  error
			conn net.Conn
		)
		defer func() {
			if errChan != nil {
				errChan <- err
			}
			if eztools.Debugging && eztools.Verbose > 1 {
				logFunc("exiting listener routine")
			}
		}()
		for {
			/*select {
				case req := <-chnSvr:
				switch req.cmd {
				case EZCOMM_CMD_END:
				}
			case*/conn, err = lstnr.Accept() //:
			if err != nil {
				logFunc("accept failed", err)
				//continue
				break
			}
			accepted(logFunc, connFunc, conn,
				[2]string{network, address})
		}
		//}
	}()
	return lstnr, nil
}
