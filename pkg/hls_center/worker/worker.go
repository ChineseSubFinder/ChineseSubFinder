package worker

import (
	"bytes"
	"context"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/hls_center/cache"
	"io"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type WorkHandler interface {
	Key(request interface{}) string
	Handle(request interface{}, w io.Writer) error
}

type WorkerServerConf struct {
	NumWorkers int
	CacheDir   string
	Worker     WorkHandler
}

type token struct{}

type WorkerServer struct {
	conf   WorkerServerConf
	cache  cache.Cache
	tokens chan token
}

func NewWorkerServer(conf WorkerServerConf) *WorkerServer {
	tokens := make(chan token, conf.NumWorkers)
	for i := conf.NumWorkers; i > 0; i-- {
		tokens <- token{}
	}
	return &WorkerServer{conf, cache.NewDirCache(conf.CacheDir), tokens}
}

func (s *WorkerServer) handler() WorkHandler {
	return s.conf.Worker
}

func (s *WorkerServer) getCachePath(r interface{}) string {
	return filepath.Join(s.conf.CacheDir, s.handler().Key(r))
}

func (s *WorkerServer) tryServeFromCache(r interface{}, w io.Writer) (bool, error) {
	data, err := s.cache.Get(context.Background(), s.handler().Key(r))
	// If error getting item, return not served with error
	if err != nil {
		return false, err
	}
	// If no item found, return not served with no error
	if data == nil {
		return false, nil
	}
	// If copying fails, return served with error
	if _, err = io.Copy(w, bytes.NewReader(data)); err != nil {
		return true, err
	}
	// Everything worked, return served with no error
	return true, nil
}

// TODO timeout & context
func (s *WorkerServer) Serve(request interface{}, w io.Writer) error {

	if served, err := s.tryServeFromCache(request, w); served || err != nil {
		if served {
			log.Debugf("Served request %v from cache", request)
		}
		if err != nil {
			log.Errorf("Error serving request from cache: %v", err)
		}
		return err
	}

	// Wait for token
	token := <-s.tokens
	defer func() {
		s.tokens <- token
	}()

	log.Debugf("Processing request %v", request)

	cw := new(bytes.Buffer)
	mw := io.MultiWriter(cw, w)
	if err := s.handler().Handle(request, mw); err != nil {
		log.Errorf("Error handling request: %v", err)
		return err
	}

	if err := s.cache.Set(context.Background(), s.handler().Key(request), cw.Bytes()); err != nil {
		log.Errorf("Error caching request: %v", err)
	}

	return nil
}
