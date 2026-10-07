package health_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adexcell/crm-ai-pet/internal/modules/health"
	"github.com/adexcell/crm-ai-pet/pkg/response"
	"github.com/go-chi/chi/v5"
)

func TestLivenessHandler(t *testing.T) {
	h := health.NewHandler(nil, nil)
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp response.ApiResponse[map[string]any]
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.OK {
		t.Errorf("expected ok to be true, got false")
	}

	if resp.Data == nil || (*resp.Data)["status"] != "alive" {
		t.Errorf("expected status alive, got %v", resp.Data)
	}
}
