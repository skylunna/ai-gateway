package api

import (
	"net/http"

	"github.com/skylunna/luner/internal/metrics"
)

func (s *RestServer) handleLiveMetrics(w http.ResponseWriter, _ *http.Request) {
	s.json(w, http.StatusOK, metrics.Snapshot())
}
