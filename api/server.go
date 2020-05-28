package api

import (
	"net/http"
	"sync/atomic"
	"time"

	. "github.com/journeymidnight/yig/error"
	"github.com/journeymidnight/yig/helper"
	"github.com/journeymidnight/yig/meta"
)

type Server struct {
	Server *http.Server
}

func (s *Server) Stop() {
	stopping = true
	helper.Logger.Info("Stopping API server")
	for {
		n := atomic.LoadInt64(&runningRequests)
		if n == 0 {
			break
		}
		time.Sleep(time.Second)
	}
	helper.Logger.Info("API Server stopped")
}

// FIXME this is ugly
var stopping = false
var runningRequests int64 = 0

type gracefulHandler struct {
	handler http.Handler
}

func (l gracefulHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if stopping {
		WriteErrorResponse(w, r, ErrMaintenance)
		return
	}
	atomic.AddInt64(&runningRequests, 1)
	defer atomic.AddInt64(&runningRequests, -1)
	l.handler.ServeHTTP(w, r)
}

func SetGracefulStopHandler(h http.Handler, _ *meta.Meta) http.Handler {
	return gracefulHandler{handler: h}
}
