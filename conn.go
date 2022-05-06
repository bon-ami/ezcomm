package main

import (
	"net"

	"gitee.com/bon-ami/eztools/v4"
)

/*const (
	EZCOMM_CMD_END = iota
	EZCOMM_CMD_SND
	EZCOMM_CMD_FIL
)

type Packs struct {
	cmd int
	str string
}

var chnSvr, chnIn, chnOut chan Packs*/

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
