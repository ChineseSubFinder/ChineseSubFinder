package base

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type StaticFileSystemBackEnd struct {
	logger     *logrus.Logger
	running    bool
	srv        *http.Server
	locker     sync.Mutex
	pathUrlMap map[string]string
}

func NewStaticFileSystemBackEnd(logger *logrus.Logger) *StaticFileSystemBackEnd {
	return &StaticFileSystemBackEnd{
		logger:     logger,
		pathUrlMap: make(map[string]string),
	}
}

// GetPathUrlMap x://电影 -- /movie_dir_0  or x://电视剧 -- /series_dir_0
func (s *StaticFileSystemBackEnd) GetPathUrlMap() map[string]string {
	return s.pathUrlMap
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

		nowUrl := "/movie_dir_" + fmt.Sprintf("%d", i)
		s.pathUrlMap[path] = nowUrl
		router.StaticFS(nowUrl, http.Dir(path))
	}
	// 添加连续剧的
	for i, path := range commonSettings.SeriesPaths {

		nowUrl := "/series_dir_" + fmt.Sprintf("%d", i)
		s.pathUrlMap[path] = nowUrl
		router.StaticFS(nowUrl, http.Dir(path))
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
	defer func() {
		s.locker.Unlock()
	}()
	s.locker.Lock()

	if s.running == false {
		s.logger.Warningln("StaticFileSystemBackEnd is not running")
		return
	}

	defer func() {
		s.pathUrlMap = make(map[string]string)
	}()

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
