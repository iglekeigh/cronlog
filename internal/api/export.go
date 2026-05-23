package api

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/user/cronlog/internal/store"
)

func (h *Handler) HandleExport(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	filter := ParseEntryFilter(r)

	entries, err := h.store.ListFiltered(r.Context(), filter, store.PageParams{Limit: 10000, Offset: 0})
	if err != nil {
		h.log.Error("export: list filtered", "error", err)
		http.Error(w, "failed to fetch entries", http.StatusInternalServerError)
		return
	}

	switch format {
	case "csv":
		h.exportCSV(w, entries)
	default:
		h.exportJSON(w, entries)
	}
}

func (h *Handler) exportJSON(w http.ResponseWriter, entries []store.Entry) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=cronlog-export.json")
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		h.log.Error("export: encode json", "error", err)
	}
}

func (h *Handler) exportCSV(w http.ResponseWriter, entries []store.Entry) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=cronlog-export.csv")

	cw := csv.NewWriter(w)
	defer cw.Flush()

	_ = cw.Write([]string{"id", "job", "status", "exit_code", "started_at", "finished_at", "output"})

	for _, e := range entries {
		finished := ""
		if e.FinishedAt != nil {
			finished = e.FinishedAt.Format(time.RFC3339)
		}
		_ = cw.Write([]string{
			strconv.FormatInt(e.ID, 10),
			e.Job,
			e.Status,
			strconv.Itoa(e.ExitCode),
			e.StartedAt.Format(time.RFC3339),
			finished,
			e.Output,
		})
	}
}
