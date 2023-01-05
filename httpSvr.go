package ezcomm

import (
	"net"
	"net/http"

	"gitee.com/bon-ami/eztools/v4"
	"github.com/gin-gonic/gin"
)

// HTTPServ is HTTP server on a relative dir
// Return value: chan from listner, wait for which to wait for the server
func HTTPServ(lst net.Listener, relativePath, rootPath string, fs http.FileSystem) chan error {
	if fs == nil {
		eztools.LogPrint("server on", rootPath, "for", relativePath)
		fs = gin.Dir(rootPath, true)
	} else {
		eztools.LogPrint("server on", relativePath)
	}
	g := gin. /*Default() */ New()
	g.StaticFS(relativePath, fs)
	ec := make(chan error, 1)
	go func() {
		ec <- g.RunListener(lst)
	}()
	return ec
}
