package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/brice-74/sensorflow/internal/log"
)

type Handlers struct {
	cfg     Signer
	service *Service
	logger  log.Logger
}

type signCSRBody struct {
	CSRPem     string `json:"csr_pem"`
	CommonName string `json:"common_name,omitempty"`
	TTLDays    int    `json:"ttl_days,omitempty"`
}

func NewHandlers(cfg Signer, service *Service, logger log.Logger) *Handlers {
	return &Handlers{cfg: cfg, service: service, logger: logger}
}

func (h *Handlers) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		h.logger.Warn(err, log.Tags{"endpoint": "healthz", "operation": "write_response"})
	}
}

func (h *Handlers) ca(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/x-pem-file")
	if _, err := w.Write(h.service.CACertPEM()); err != nil {
		h.logger.Warn(err, log.Tags{"endpoint": "ca", "operation": "write_response"})
	}
}

func (h *Handlers) signClientCSR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body signCSRBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	resp, err := h.service.SignClientCSR(SignCSRRequest{
		CSRPem:     body.CSRPem,
		CommonName: body.CommonName,
		TTLDays:    body.TTLDays,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if isBadRequest(err) {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error(err, log.Tags{"endpoint": "sign_client_csr", "operation": "encode_response"})
	}
}

func isBadRequest(err error) bool {
	if err == nil {
		return false
	}
	if strings.Contains(err.Error(), "required") ||
		strings.Contains(err.Error(), "invalid") ||
		strings.Contains(err.Error(), "parse CSR") {
		return true
	}
	var target *json.SyntaxError
	if errors.As(err, &target) {
		return true
	}
	return false
}

func registerRoutes(mux *http.ServeMux, h *Handlers) {
	mux.HandleFunc("/healthz", h.healthz)
	mux.HandleFunc("/v1/pki/ca", h.ca)
	mux.HandleFunc("/v1/pki/sign/client-csr", h.signClientCSR)
}
