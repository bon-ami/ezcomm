package ezcomm

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"gitee.com/bon-ami/eztools/v5"
	"github.com/gin-gonic/gin"
)

// HTTPSvr Serv=serve Shut=shutdown
type HTTPSvr struct {
	svr http.Server
	rt  *gin.Engine
}

// MakeHTTPSvr makes a HTTPSvr
func MakeHTTPSvr() *HTTPSvr {
	return &HTTPSvr{rt: gin. /*Default() */ New()}
}

// FS sets a static file system
func (svr *HTTPSvr) FS(relativePath, rootPath string, fs http.FileSystem) {
	if fs == nil {
		if eztools.Debugging && eztools.Verbose > 1 {
			eztools.LogPrint("server on path", rootPath, "for relative path", relativePath)
		}
		fs = gin.Dir(rootPath, true)
	} else {
		if eztools.Debugging && eztools.Verbose > 1 {
			eztools.LogPrint("server on relative path", relativePath)
		}
	}
	svr.rt.StaticFS(relativePath, fs)
}

// Serve http.Serve with handler set by FS()
// Return value: error from http.Serve()
func (svr *HTTPSvr) Serve(lst net.Listener) chan error {
	svr.svr.Handler = svr.rt
	ec := make(chan error, 1)
	go func() {
		if err := svr.svr.Serve(lst); err != nil && !errors.Is(err, http.ErrServerClosed) {
			ec <- err
		} else {
			ec <- nil
		}
	}()
	return ec
}

// Shutdown gracefully shuts down the server
// Return value: context.DeadlineExceeded if timeout
func (svr *HTTPSvr) Shutdown(timeout time.Duration) error {
	/*// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit*/

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	if timeout < 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return svr.svr.Shutdown(ctx)
}
