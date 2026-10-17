package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"example.com/property-access-audit/internal/audit"
)

type auditRequest struct {
	Grants []audit.Grant      `json:"grants"`
	Data   audit.PropertyData `json:"data"`
}

type auditResponse struct {
	Report audit.Report    `json:"report"`
	PDF    json.RawMessage `json:"pdf"`
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	client := audit.NewClient(key)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /audits", handleAudit(client))
	addr := envOr("ADDR", ":8080")
	log.Printf("property access audit listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func handleAudit(client *audit.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input auditRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid audit request"})
			return
		}

		keys, err := client.ListKeys(r.Context())
		if err != nil {
			writeClientError(w, err)
			return
		}
		report := audit.BuildReport(keys, input.Grants, time.Now())
		pdf, err := client.GeneratePDF(r.Context(), audit.Markdown(report))
		if err != nil {
			writeClientError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, auditResponse{Report: report, PDF: pdf})
	}
}

func writeClientError(w http.ResponseWriter, err error) {
	var apiErr *audit.InfraiError
	if errors.As(err, &apiErr) && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
		writeJSON(w, apiErr.HTTPStatus, map[string]string{"error": apiErr.Code, "message": apiErr.Message})
		return
	}
	log.Printf("Infrai request failed: %v", err)
	writeJSON(w, http.StatusBadGateway, map[string]string{"error": "upstream request failed"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
