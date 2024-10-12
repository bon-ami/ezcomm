package ezcomm

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"gitee.com/bon-ami/eztools/v6"
)

// FuncLog is log function
type FuncLog func(...any)

// FuncConn is run when Connected.
// addr=address info.
//
//	[0]: parsed local
//	[1]: remote
//	[2]: requested protocol
//	[3]: requested address
//
// close(chan[...]) upon exiting!
//
//	[0]: to ezcomm
//	[1]: from ezcomm
type FuncConn func(addr [4]string, chn [2]chan RoutCommStruc)

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

// RoutCommStruc is communication struct between EZComm and user
type RoutCommStruc struct {
	// ReqAddr is address from user
	ReqAddr string
	// Act action
	Act int
	// PeerUDP is filled, and required for FlowChnSnd by EZ Comm
	PeerUDP *net.UDPAddr
	// PeerTCP is filled, but not required by EZ Comm
	PeerTCP net.Addr
	// Data is I/O raw message
	Data []byte
	// Resp for SvrTcp and flow only
	Resp chan RoutCommStruc
	Err  error
}

func replyWtErr(_ FuncLog, comm RoutCommStruc, err error, chn chan RoutCommStruc) {
	//logFunc("UDP", cmd.PeerUdp, n, err)
	comm.Err = err
	chn <- comm
	//logFunc("UDP sent", comm)
}

func connectedUDPNet(logFunc FuncLog, chn [2]chan RoutCommStruc,
	conn *net.UDPConn, lcl string) {
	var errRet error
	defer func() {
		chn[1] <- RoutCommStruc{
			Act: FlowChnEnd,
			Err: errRet,
		}
		if eztools.Debugging && eztools.Verbose > 1 {
			logFunc("exiting routine", lcl)
		}
	}()
	floodRecs := make(map[string][2]int64)
	for {
		buf := make([]byte, FlowRcvLen)
		if eztools.Debugging && eztools.Verbose > 2 {
			logFunc("receiving UDP", conn.LocalAddr())
		}
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil &&
			(errors.Is(err, io.EOF) ||
				errors.Is(err, net.ErrClosed)) {
			break
		}
		if err != nil {
			errRet = err
			break
		}
		if addr == nil {
			// such as an err buf is smaller than data
			continue
		}
		if eztools.Debugging && eztools.Verbose > 2 {
			logFunc("received UDP from", n, addr, err, buf[:n])
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
		isFile, isEnd := IsDataFile(buf[:n])
		if !isFile || isEnd {
			//logFunc("checking flood", peerHost,
			//floodRecs[peerHost])
			ok, recs := floodCtrl(floodRecs[peerHost],
				addr.String())
			if ok {
				if eztools.Debugging &&
					eztools.Verbose > 2 {
					logFunc("flooding", peerAddr)
				}
				continue
			}
			floodRecs[peerHost] = recs
		}
		comm.Data = buf[:n]
		comm.PeerUDP = addr
		chn[1] <- comm
	}
}

// ConnectedUDP works for UDP, when remote can change
//
//	It blocks.
//
// chn[1] -> caller: FlowChnRcv, FlowChnSnd, FlowChnEnd
// chn[0] <- caller: FlowChnSnd, FlowChnEnd
func ConnectedUDP(logFunc FuncLog, chn [2]chan RoutCommStruc, conn *net.UDPConn) {
	defer conn.Close()
	lcl := conn.LocalAddr().String()
	if eztools.Debugging && eztools.Verbose > 1 {
		logFunc("entering udp", conn.LocalAddr().Network(), lcl)
		defer func() {
			logFunc("exiting udp", lcl)
		}()
	}
	go connectedUDPNet(logFunc, chn, conn, lcl)
	if eztools.Debugging && eztools.Verbose > 2 {
		logFunc("listening on UDP", chn)
	}
	var err error
	for {
		if eztools.Debugging && eztools.Verbose > 2 {
			logFunc("command udp", chn, "waiting")
		}
		cmd := <-chn[0]
		if eztools.Debugging && eztools.Verbose > 2 {
			logFunc("command udp", chn, cmd)
		}
		switch cmd.Act {
		case FlowChnSnd:
			/*if eztools.Debugging && eztools.Verbose > 2 {
				logFunc("sending", cmd)
			}*/
			// TODO: confirm to anti-flood
			if cmd.PeerUDP == nil {
				err = eztools.ErrIncomplete
			} else {
				_, err = conn.WriteToUDP(cmd.Data, cmd.PeerUDP)
			}
			replyWtErr(logFunc, cmd, err, chn[1])
		case FlowChnEnd:
			//logFunc("exiting", lcl)
			return
		}
	}
}

// floodCtrl check for flood in a connection
// Parameters: getRecs[0]=previous start of flood check;
// [1]=number of packets within a second
// return values: flooding; updated data for getRecs
func floodCtrl(floodRecs [2]int64,
	peer string) (bool, [2]int64) {
	if AntiFlood.Limit < 0 {
		return false, [2]int64{0, 0}
	}
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
	//eztools.Log(curr, AntiFlood.Limit, floodRecs)
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

func rcvFrom1Peer(logFunc FuncLog, conn net.Conn, chn chan RoutCommStruc,
	localAddr string, peer net.Addr, peerHost, peerAddr string) {
	if eztools.Debugging && eztools.Verbose > 1 {
		logFunc("entering routine",
			conn.LocalAddr().Network(), localAddr)
	}
	var (
		floodRecs [2]int64
		errRet    error
	)
	defer func() {
		if eztools.Debugging && eztools.Verbose > 1 {
			logFunc("exiting TCP/UDP", localAddr,
				"routine peer", peerAddr)
		}
		chn <- RoutCommStruc{
			Act: FlowChnEnd,
			Err: errRet,
		}
	}()
	for {
		buf := make([]byte, FlowRcvLen)
		if eztools.Debugging && eztools.Verbose > 2 {
			logFunc("receiving TCP/UDP", localAddr)
		}
		n, err := conn.Read(buf)
		if err != nil {
			if !errors.Is(err, io.EOF) &&
				!errors.Is(err, net.ErrClosed) {
				// reading on a closed TCP connection also here
				errRet = err
			}
			break
		}
		if eztools.Debugging && eztools.Verbose > 2 {
			logFunc("received TCP/UDP", localAddr, n, "from", peerAddr, err, buf[:n])
		}
		comm := RoutCommStruc{
			Act:     FlowChnRcv,
			PeerTCP: peer,
		}
		floodChk := func() bool {
			isFile, isEnd := IsDataFile(comm.Data)
			if isFile && !isEnd {
				return false
			}
			var ok bool
			ok, floodRecs = floodCtrl(
				floodRecs, peerHost)
			if ok {
				if eztools.Debugging && eztools.Verbose > 0 {
					logFunc("flooding", peerHost)
				}
				return true
			}
			return false
		}

		comm.Data = make([]byte, n)
		copy(comm.Data, buf[:n])
		if floodChk() {
			errRet = eztools.ErrAccess
			break
		}
		chn <- comm
	}
}

// Connected1Peer works for TCP and UDP, when remote does not change
// Parameters:
//
//	logFunc for logging
//	connFunc is callback function upon entrance of this function.
//	  It blocks the routine.
//	  If flood detected, it is called with local and remote addresses,
//	  requested protocol and address, and nil channels
//	  eztools.ErrAccess is sent to chan[1].
//	  Network failure after the socket is closed cannot be matched,
//	  if not net.ErrClosed or io.EOF,
//	  so it is sent to chan[1] directly.
//	conn is the connection
//	addrReq is address user requested, and varies between local/remote,
//	  when user creates a server (listen) or a client
//	 -> caller: connFunc(), FlowChnRcv, FlowChnSnd, FlowChnEnd
//	 <- caller: FlowChnSnd, FlowChnEnd // this may block routine from exiting,
//	            if too much incoming traffic not read
func Connected1Peer(logFunc FuncLog, connFunc FuncConn,
	conn net.Conn, addrReq [2]string) {
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
		go connCB()
		return
	}
	for i := range chn {
		chn[i] = make(chan RoutCommStruc, FlowComLen)
	}
	go connCB()
	go rcvFrom1Peer(logFunc, conn, chn[1], localAddr, peer, peerHost, peerAddr)
	if eztools.Debugging && eztools.Verbose > 0 {
		logFunc("listening on TCP/UDP", chn)
	}
	for {
		if eztools.Debugging && eztools.Verbose > 2 {
			logFunc("command tcp/udp", chn, "waiting")
		}
		cmd := <-chn[0]
		if eztools.Debugging && eztools.Verbose > 2 {
			logFunc("command tcp/udp", chn, peerAddr, cmd)
		}
		switch cmd.Act {
		case FlowChnSnd:
			cmd.PeerTCP = peer
			// TODO: confirm to anti-flood
			_, err := conn.Write(cmd.Data)
			replyWtErr(logFunc, cmd, err, chn[1])
		case FlowChnEnd:
			if eztools.Debugging && eztools.Verbose > 1 {
				logFunc("exiting TCP/UDP connection routine",
					localAddr, "peer", peerAddr)
			}
			return
		}
	}
}

// Client is mainly for TCP and Unix(not -gram). It ends immediately.
//
//	lstnFunc() needs to handle procedures afterwards.
//	returned conn needs to Close() by user.
//
// UDP, IP and Unixgram will be tricky to track peer information,
//
//	and preferrably use Listen*()
func Client(logFunc FuncLog, connFunc FuncConn,
	network, rmtAddr string,
	lstnFunc func(logFunc FuncLog, connFunc FuncConn,
		conn net.Conn, rmtAddr [2]string)) (
	conn net.Conn, err error) {
	conn, err = net.Dial(network, rmtAddr)
	if err != nil {
		return
	}
	//defer conn.Close()
	if lstnFunc != nil {
		go lstnFunc(logFunc, connFunc, conn,
			[2]string{network, rmtAddr})
	}
	return
}

// ListenIP listens to IP. It ends immediately.
// Parameters:
//
//	network is socket type, "ip", "ip4" or "ip6"
//	address is for local, [IP or DN][:protocol number or name]
func ListenIP(network, address string) (*net.IPConn, error) {
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

// ListenUnixgram listens to unixgram. It ends immediately.
// Parameters:
//
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

// ListenUDP listens to UDP. It ends immediately.
//
//	*net.UDPConn needs to handle procedures afterwards.
//
// Parameters:
//
//	network is socket type, "udp", "udp4" or "udp6"
//	address is for local, [IP or DN][:port number or name]
func ListenUDP(network, address string) (*net.UDPConn, error) {
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

// ListenTCP listens to TCP. It ends immediately.
// Simply close the listener to stop it.
//
//	Parameters:
//	network is socket type, "tcp", "tcp4" or "tcp6"
//	accepted() needs to handle procedures for incoming connections.
//	  can be nil.
//	  It blocks listening routine, so that errChan gets feedback
//	  always before accepted().
//	connFunc is for accepted() only.
//	errChan sends Accept errors
func ListenTCP(logFunc FuncLog, connFunc FuncConn,
	network, address string, accepted func(FuncLog,
		FuncConn, net.Conn, [2]string), errChan chan error) (
	net.Listener, error) {
	if logFunc == nil {
		logFunc = func(...any) {}
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
	if accepted == nil {
		return lstnr, nil
	}
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
			conn, err = lstnr.Accept()
			if err != nil {
				if eztools.Debugging && eztools.Verbose > 0 {
					logFunc("accept failed", err)
				}
				//continue
				break
			}
			go accepted(logFunc, connFunc, conn,
				[2]string{network, address})
		}
	}()
	return lstnr, nil
}
