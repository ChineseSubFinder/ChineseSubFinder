package backend

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/frontend/dist"
	"github.com/allanpk716/ChineseSubFinder/internal/backend/routers"
	"github.com/allanpk716/ChineseSubFinder/internal/backend/ws_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/cron_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

// StartBackEnd 开启后端的服务器
func StartBackEnd(fileDownloader *file_downloader.FileDownloader, httpPort int, cronHelper *cron_helper.CronHelper) {

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	engine := gin.Default()
	// 默认所有都通过
	engine.Use(cors.Default())
	v1Router := routers.InitRouter(fileDownloader, engine, cronHelper)
	defer func() {
		v1Router.Close()
	}()

	engine.GET("/", func(c *gin.Context) {
		c.Header("content-type", "text/html;charset=utf-8")
		c.String(http.StatusOK, string(dist.SpaIndexHtml))
	})
	engine.StaticFS(dist.SpaFolderJS, dist.Assets(dist.SpaFolderName+dist.SpaFolderJS, dist.SpaJS))
	engine.StaticFS(dist.SpaFolderCSS, dist.Assets(dist.SpaFolderName+dist.SpaFolderCSS, dist.SpaCSS))
	engine.StaticFS(dist.SpaFolderFonts, dist.Assets(dist.SpaFolderName+dist.SpaFolderFonts, dist.SpaFonts))
	engine.StaticFS(dist.SpaFolderIcons, dist.Assets(dist.SpaFolderName+dist.SpaFolderIcons, dist.SpaIcons))
	engine.StaticFS(dist.SpaFolderImages, dist.Assets(dist.SpaFolderName+dist.SpaFolderImages, dist.SpaImages))

	engine.Any("/api", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/")
	})

	hub := ws_helper.NewHub()
	go hub.Run()
	defer func() {
		hub.Clear()
	}()
	engine.GET("/ws", func(context *gin.Context) {
		ws_helper.ServeWs(fileDownloader.Log, hub, context.Writer, context.Request)
	})

	// listen and serve on 0.0.0.0:8080(default)
	fileDownloader.Log.Infoln("Try Start Server At Port", httpPort)
	err := engine.Run(":" + fmt.Sprintf("%d", httpPort))
	if err != nil {
		fileDownloader.Log.Errorln("Start Server At Port", httpPort, "Error", err)
	}
}
