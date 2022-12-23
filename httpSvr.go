package ezcomm

import (
	"net"

	"github.com/gin-gonic/gin"
)

func HttpServ(dir string, lst net.Listener) error {
	g := gin.New()
	g.Static("", dir)
	return g.RunListener(lst)
}
