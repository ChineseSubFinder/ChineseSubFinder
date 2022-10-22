package backend

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/allanpk716/ChineseSubFinder/pkg/settings"

	"github.com/allanpk716/ChineseSubFinder/frontend/dist"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/sirupsen/logrus"

	"github.com/allanpk716/ChineseSubFinder/pkg/logic/cron_helper"
)

type BackEnd struct {
	logger        *logrus.Logger
	settings      *settings.Settings
	cronHelper    *cron_helper.CronHelper
	httpPort      int
	running       bool
	srv           *http.Server
	locker        sync.Mutex
	restartSignal chan interface{}
}

func NewBackEnd(logger *logrus.Logger, settings *settings.Settings, cronHelper *cron_helper.CronHelper, httpPort int, restartSignal chan interface{}) *BackEnd {
	return &BackEnd{logger: logger, settings: settings, cronHelper: cronHelper, httpPort: httpPort, restartSignal: restartSignal}
}

func (b *BackEnd) start() {

	defer b.locker.Unlock()
	b.locker.Lock()

	if b.running == true {
		b.logger.Debugln("Http Server is already running")
		return
	}
	b.running = true

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	engine := gin.Default()
	// 默认所有都通过
	engine.Use(cors.Default())
	cbBase, v1Router := InitRouter(b.settings, engine, b.cronHelper, b.restartSignal)

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

	// listen and serve on 0.0.0.0:8080(default)
	b.srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", b.httpPort),
		Handler: engine,
	}
	go func() {
		b.logger.Infoln("Try Start Http Server At Port", b.httpPort)
		if err := b.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			b.logger.Errorln("Start Server Error:", err)
		}
		defer func() {
			cbBase.Close()
			v1Router.Close()
		}()
	}()
}

func (b *BackEnd) Restart() {

	stopFunc := func() {

		b.locker.Lock()
		defer func() {
			b.locker.Unlock()
		}()
		if b.running == false {
			b.logger.Debugln("Http Server is not running")
			return
		}
		b.running = false

		exitOk := make(chan interface{}, 1)
		defer close(exitOk)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		go func() {
			if err := b.srv.Shutdown(ctx); err != nil {
				b.logger.Errorln("Http Server Shutdown:", err)
			}
			exitOk <- true
		}()
		select {
		case <-ctx.Done():
			b.logger.Warningln("Http Server Shutdown timeout of 5 seconds.")
		case <-exitOk:
			b.logger.Infoln("Http Server Shutdown Successfully")
		}
		b.logger.Infoln("Http Server Shutdown Done.")
	}

	for {
		select {
		case <-b.restartSignal:
			{
				stopFunc()
				b.start()
			}
		}
	}
}
