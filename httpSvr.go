package ezcomm

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"gitee.com/bon-ami/eztools/v6"
	"github.com/gin-gonic/gin"
)

// HTTPSvr Serv=serve Shut=shutdown
type HTTPSvr struct {
	svr http.Server
	rt  *gin.Engine
}

// MakeHTTPSvr makes a HTTPSvr
func MakeHTTPSvr() *HTTPSvr {
	if !eztools.Debugging {
		gin.SetMode(gin.ReleaseMode)
	}
	return &HTTPSvr{rt: gin. /*Default() */ New()}
}

// FS sets a static file system
// Parameters: fs OR rootPath: to use a filesystem or gin's default
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

const (
	// HTTPSvrHTMLContentType ContentType of HTML for HTTPSvr
	HTTPSvrHTMLContentType = "text/html; charset=utf-8"
	// HTTPSvrBodyString string as body for HTTPSvr
	HTTPSvrBodyString = iota
	// HTTPSvrBodyHTML html as body for HTTPSvr
	// this is same as HTTPSvrBodyString,
	// in addition, it sets ContentType to HTTPSvrHTMLContentType
	HTTPSvrBodyHTML
	// HTTPSvrBodyJSON json as body for HTTPSvr
	HTTPSvrBodyJSON
)

// HTTPSvrBody body for HTTPSvr
type HTTPSvrBody struct {
	// Tp type, one of HTTPSvrBodyString, HTTPSvrBodyHTML
	// and HTTPSvrBodyJSON
	Tp int
	// Str for HTTPSvrBodyString and HTTPSvrBodyHTML
	Str string
	// JSON for HTTPSvrBodyJSON, map[string]any
	JSON gin.H
}

// HTTPSvrProcFunc method proc func type for HTTPSvr
//
//	Parameters:
//
// remote IP
// request
// func to get values by keys
//
//	Return values:
//	response code, body, headers. 204 is default if fun is nil.
type HTTPSvrProcFunc func(string, *http.Request,
	func(string) string) (int, HTTPSvrBody, map[string]string)

func (svr *HTTPSvr) methodsProc(c *gin.Context, fun HTTPSvrProcFunc) {
	//buf, err := c.GetRawData()
	code := http.StatusNoContent
	var body HTTPSvrBody
	if fun != nil {
		var pS map[string]string
		code, body, pS = fun(c.ClientIP(), c.Request, c.PostForm)
		for k, v := range pS {
			/*k, v, e := pS.GetNMove()
			if e != nil {
				break
			}*/
			c.Header(k, v)
		}
	}
	switch body.Tp {
	case HTTPSvrBodyHTML:
		c.Header("Content-Type", HTTPSvrHTMLContentType)
		fallthrough
	case HTTPSvrBodyString:
		c.String(code, body.Str)
	case HTTPSvrBodyJSON:
		c.JSON(code, body.JSON)
	}
}

// GET sets a GET handler
func (svr *HTTPSvr) GET(relativePath string, fun HTTPSvrProcFunc) {
	svr.rt.GET(relativePath, func(c *gin.Context) {
		svr.methodsProc(c, fun)
	})
}

// POST shows a text input
func (svr *HTTPSvr) POST(relativePath string, fun HTTPSvrProcFunc) {
	svr.rt.POST(relativePath, func(c *gin.Context) {
		svr.methodsProc(c, fun)
	})
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
