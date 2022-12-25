package ezcomm

import (
	"net"

	"github.com/gin-gonic/gin"
)

func HttpServ(lst net.Listener, dir string) error {
	g := gin. /*Default() */ New()
	g.StaticFS("", gin.Dir(dir, true))
	//g.StaticFS("", http.Dir(dir))
	ec := make(chan error, 1)
	go func() {
		ec <- g.RunListener(lst)
	}()
	<-ec
	return nil
}
