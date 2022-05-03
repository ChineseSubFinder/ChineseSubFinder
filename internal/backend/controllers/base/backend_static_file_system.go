package base

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"net/http"
	"sync"
	"time"
)

type StaticFileSystemBackEnd struct {
	logger  *logrus.Logger
	running bool
	srv     *http.Server
	locker  sync.Mutex
}

func NewStaticFileSystemBackEnd(logger *logrus.Logger) *StaticFileSystemBackEnd {
	return &StaticFileSystemBackEnd{
		logger: logger,
	}
}

func (s *StaticFileSystemBackEnd) Start(commonSettings *settings.CommonSettings) {
	defer s.locker.Unlock()
	s.locker.Lock()

	if s.running == true {
		s.logger.Warningln("StaticFileSystemBackEnd is already running")
		return
	}

	s.running = true

	router := gin.Default()

	// 添加电影的
	for i, path := range commonSettings.MoviePaths {
		router.StaticFS("/movie_dir_"+fmt.Sprintf("%d", i), http.Dir(path))
	}
	// 添加连续剧的
	for i, path := range commonSettings.SeriesPaths {
		router.StaticFS("/series_dir_"+fmt.Sprintf("%d", i), http.Dir(path))
	}
	s.srv = &http.Server{
		Addr:    ":" + commonSettings.LocalStaticFilePort,
		Handler: router,
	}
	go func() {
		// service connections
		s.logger.Infoln("Listening and serving HTTP on port " + commonSettings.LocalStaticFilePort)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Errorln("StaticFileSystemBackEnd listen:", err)
		}
	}()
}

func (s *StaticFileSystemBackEnd) Stop() {
	defer s.locker.Unlock()
	s.locker.Lock()

	if s.running == false {
		s.logger.Warningln("StaticFileSystemBackEnd is not running")
		return
	}

	s.running = false

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		s.logger.Errorln("Server Shutdown:", err)
	}
	select {
	case <-ctx.Done():
		s.logger.Warningln("timeout of 5 seconds.")
	}
	s.logger.Infoln("Static File System Server exiting")
}
