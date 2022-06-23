package ezcomm

import (
	"errors"
	"io"
	"net"

	"gitee.com/bon-ami/eztools/v4"
)

type Guis interface {
	/*// GuiSetGlbPrm is run at the beginning,
	//   to initialize following for ezcomm.
	//   Ver, Bld, GuiConnected, GuiEnded, GuiLog, GuiRcv, GuiSnt
	GuiSetGlbPrm
	// Run is run at the end to handle UI in the main thread*/
	Run(Ver, Bld string)
	Log(bool, ...any)
	Rcv(RoutCommStruc)
	Snt(RoutCommStruc)
	Connected(string, string)
	Ended(RoutCommStruc)
}

var (
	ChanComm [2]chan RoutCommStruc
	// Snd2Slc is for UDP peer display
	Snd2Slc [2][]string
	// SndMap is for UDP peer match
	SndMap [2]map[string]struct{}
	RecMap map[string][]string
	RecSlc []string
	RcvMap map[string]struct{}
	PeeMap map[string]chan RoutCommStruc
	gui    Guis
)

func SetGui(g Guis) {
	gui = g
}

type RoutCommStruc struct {
	// Act=FlowChn* (>FlowChnMax)
	Act     int
	PeerUdp *net.UDPAddr
	PeerTcp net.Addr
	Data    string
	Resp    chan RoutCommStruc
	Err     error
}

func ConnectedUdp(conn *net.UDPConn) {
	buf := make([]byte, FlowRcvLen)
	lcl := conn.LocalAddr().String()
	go func() {
		if gui != nil && eztools.Debugging && eztools.Verbose > 1 {
			defer gui.Log(true, "exiting routine", lcl)
		}
		for {
			//GuiLog(true, "receiving UDP", conn.LocalAddr())
			n, addr, err := conn.ReadFromUDP(buf)
			//GuiLog(true, "received UDP", n, addr, err)
			if err != nil &&
				(errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed)) {
				return
			}
			comm := RoutCommStruc{
				Act: FlowChnRcv,
				Err: err,
			}
			if err == nil {
				comm.Data = string(buf[:n])
				comm.PeerUdp = addr
			}
			//chanComm[1] <- comm
			if gui != nil {
				gui.Rcv(comm)
			}
		}
	}()
	//GuiLog(true, "listening on UDP", ChanComm[0])
	for {
		cmd := <-ChanComm[0]
		switch cmd.Act {
		case FlowChnSnd:
			/*if eztools.Debugging && eztools.Verbose > 2 {
				GuiLog(true, "sending", cmd)
			}*/
			_, err := conn.WriteToUDP([]byte(cmd.Data), cmd.PeerUdp)
			/*chanComm[1] <- FlowCommStruc{
				Act: FlowChnSnd,
				Err: err,
			}*/
			comm := cmd
			cmd.Err = err
			if gui != nil {
				gui.Snt(comm)
			}
			//GuiLog(true, "UDP sent", comm)
		case FlowChnEnd:
			for i := range ChanComm {
				ChanComm[i] = nil
			}
			conn.Close()
			if gui != nil && eztools.Debugging && eztools.Verbose > 1 {
				gui.Log(true, "exiting", lcl)
			}
			return
		}
	}
}

func ListeningTcp(lstnr net.Listener) {
	if gui != nil && eztools.Debugging && eztools.Verbose > 1 {
		defer gui.Log(true, "exiting server", lstnr.Addr().String())
	}
	for {
		cmd := <-ChanComm[0]
		switch cmd.Act {
		case FlowChnEnd:
			lstnr.Close()
			for i := range ChanComm {
				ChanComm[i] = nil
			}
			return
		}
	}
}

func ConnectedTcp(conn net.Conn) {
	chn := make(chan RoutCommStruc, FlowComLen)
	peer := conn.RemoteAddr()
	PeeMap[peer.String()] = chn
	if gui != nil {
		gui.Connected(conn.LocalAddr().String(), peer.String())
	}
	buf := make([]byte, FlowRcvLen)
	go func() {
		if gui != nil && eztools.Debugging && eztools.Verbose > 1 {
			defer gui.Log(true, "exiting routine peer", peer.String())
		}
		for {
			n, err := conn.Read(buf)
			comm := RoutCommStruc{
				Act:     FlowChnRcv,
				PeerTcp: peer,
				Err:     err,
			}
			if err == nil {
				comm.Data = string(buf[:n])
			} else {
				if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
					comm.Act = FlowChnEnd
					if gui != nil {
						gui.Ended(comm)
					}
					break
				}
			}
			if gui != nil {
				gui.Rcv(comm)
			}
		}
	}()
	for {
		cmd := <-chn
		switch cmd.Act {
		case FlowChnSnd:
			_, err := conn.Write([]byte(cmd.Data))
			/*chanComm[1] <- FlowCommStruc{
				Act: FlowChnSnd,
				Err: err,
			}*/
			comm := cmd
			comm.Err = err
			comm.PeerTcp = peer
			if gui != nil {
				gui.Snt(comm)
			}
		case FlowChnEnd:
			conn.Close()
			if gui != nil && eztools.Debugging && eztools.Verbose > 1 {
				gui.Log(true, "exiting peer", peer.String())
			}
			return
		}
	}
}

// Client is mainly for TCP and Unix(not -gram). It ends immediately.
//   fun() needs to handle procedures afterwards.
// UDP, IP and Unixgram will be tricky to track peer information,
//   and preferrably use Listen*()
func Client(network, address string,
	fun func(net.Conn)) (net.Conn, error) {
	/*if fun == nil {
		eztools.LogFatal("no function to handle server")
	}*/
	conn, err := net.Dial(network, address)
	if err != nil {
		return conn, err
	}
	//defer conn.Close()
	if fun != nil {
		go func() {
			fun(conn)
		}()
	}
	return conn, nil
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
	if len(address) > 0 {
		addr, err = net.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
	}
	return net.ListenUDP(network, addr)
}

// ListenTcp listens to TCP. It ends immediately.
//   accepted() needs to handle procedures for incoming connections.
// Parameters:
//	network is socket type, "tcp", "tcp4", "tcp6",
//		"unix" or "unixpacket"
//	fun handles incoming connections for TCP/unix and all connections for UDP
//	errChan sends Accept errors for TCP
func ListenTcp(network, address string, accepted func(net.Conn),
	errChan chan error) (net.Listener, error) {
	if accepted == nil {
		eztools.LogFatal("no function to handle server")
	}
	//eztools.Log("to serve")
	lstnr, err := net.Listen(network, address)
	if err != nil {
		return lstnr, err
	}
	//eztools.Log("serving", lstnr)
	//defer lstnr.Close()
	go func() {
		for {
			/*select {
				case req := <-chnSvr:
				switch req.cmd {
				case EZCOMM_CMD_END:
				}
			case*/conn, err := lstnr.Accept() //:
			if err != nil {
				/*eztools.Log("accept failed", err)
				continue*/
				if errChan != nil {
					errChan <- err
				}
				break
			}
			go accepted(conn)
		}
		//}
	}()
	return lstnr, nil
}
