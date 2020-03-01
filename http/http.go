package http

import (
	"fmt"
	"net/http"

	"github.com/didi/falcon-log-agent/common/utils"

	"github.com/didi/falcon-log-agent/common/g"
	"github.com/didi/falcon-log-agent/strategy"
	"github.com/didi/falcon-log-agent/worker"

	"github.com/gin-gonic/gin"
)

// Start http api
func Start() {
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, "ok")
	})
	router.GET("/strategy", func(c *gin.Context) {
		c.JSON(http.StatusOK, strategy.GetListAll())
	})

	router.GET("/cached", func(c *gin.Context) {
		c.String(http.StatusOK, worker.GetCachedAll())
	})

	router.POST("/check", func(c *gin.Context) {
		log := c.PostForm("log")
		c.JSON(http.StatusOK, CheckLogByStrategy(log))
	})

	ip, err := utils.LocalIP()
	if err != nil {
		ip = "127.0.0.1"
	}
	router.Run(fmt.Sprintf("%s:%d", ip, g.Conf().Http.HTTPPort))
}
