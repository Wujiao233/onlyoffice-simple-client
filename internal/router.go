package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"wwwj.dev/wujiao/onlyoffice-simple-client/internal/apis"
)

func InitRouter() *gin.Engine {
	router := gin.New()
	router.GET("/ping", apis.Ping)

	pageGroup := router.Group("/files")
	pageGroup.Use(apis.CheckAuth())
	pageGroup.StaticFS("/", http.Dir("./data"))

	router.StaticFS("/static", http.Dir("./static"))
	router.LoadHTMLGlob("./templates/*")

	router.POST("/callback", apis.Callback)
	router.POST("/upload", apis.Upload)
	router.POST("/del", apis.Del)
	router.GET("/", apis.Index)
	router.GET("/genConfig", apis.GenConfig)

	router.GET("fileList", apis.FileList)

	//router.GET("/hivetest", hive.HdfsTest)
	return router
}
