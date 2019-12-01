package matcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	clairerror "github.com/quay/clair/v4/clair-error"
	je "github.com/quay/claircore/pkg/jsonerr"
)

const (
	// VulnerabilityReportAPIPath is the http path for accessing vulnerability_report
	VulnerabilityReportAPIPath = "/api/v1/vulnerability_report/"
)

// MatchHandler is an HTTP handler for matching IndexReports to VulnerabilityReports
func MatchHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			resp := &je.Response{
				Code:    "method-not-allowed",
				Message: "endpoint only allows GET",
			}
			je.Error(w, resp, http.StatusMethodNotAllowed)
			return
		}

		manifestHash := strings.TrimPrefix(r.URL.Path, VulnerabilityReportAPIPath)
		if manifestHash == "" {
			resp := &je.Response{
				Code:    "bad-request",
				Message: "malformed path. provide a single manifest hash",
			}
			je.Error(w, resp, http.StatusBadRequest)
			return
		}

		report, err := service.Match(context.Background(), manifestHash)
		if err != nil {
			if errors.Is(err, &clairerror.ErrIndexReportNotFound{}) {
				resp := &je.Response{
					Code:    "not-found",
					Message: fmt.Sprintf("index report for manifest %s not found", manifestHash),
				}
				je.Error(w, resp, http.StatusNotFound)
				return
			}
			resp := &je.Response{
				Code:    "match-error",
				Message: fmt.Sprintf("failed to start scan: %v", err),
			}
			je.Error(w, resp, http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(report)
		if err != nil {
			resp := &je.Response{
				Code:    "encoding-error",
				Message: fmt.Sprintf("failed to encode vulnerability report: %v", err),
			}
			je.Error(w, resp, http.StatusInternalServerError)
			return
		}

		return
	}
}