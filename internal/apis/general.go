package apis

import (
	"github.com/gin-gonic/gin"
	"wwwj.dev/wujiao/onlyoffice-simple-client/internal/utils/config"
	"wwwj.dev/wujiao/onlyoffice-simple-client/internal/utils/resp"
)

func Ping(c *gin.Context) {
	resp.OK(c, "pong")
}

func Index(c *gin.Context) {
	c.HTML(200, "index.html", gin.H{
		"documentserver": config.Conf.Onlyoffice.Host,
	})
}
