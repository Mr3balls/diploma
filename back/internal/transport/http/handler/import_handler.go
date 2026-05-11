package handler

import (
	"net/http"

	"esports-backend/internal/apperror"
	"esports-backend/internal/service"

	"github.com/go-chi/chi/v5"
)

type ImportHandler struct{ deps Deps }

func NewImportHandler(deps Deps) *ImportHandler { return &ImportHandler{deps: deps} }

type connectSheetRequest struct {
	SheetURL      string `json:"sheet_url" validate:"required,url"`
	WorksheetName string `json:"worksheet_name" validate:"required"`
}

type confirmImportRequest struct {
	BatchID string `json:"batch_id" validate:"required,uuid"`
}

func (h *ImportHandler) ConnectSheet(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	tournamentID := chi.URLParam(r, "id")
	var req connectSheetRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.deps.Validate.Struct(req); err != nil {
		writeError(w, err)
		return
	}
	result, err := h.deps.Imports.Connect(r.Context(), actorUserID, tournamentID, service.ConnectSheetInput{SheetURL: req.SheetURL, WorksheetName: req.WorksheetName})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *ImportHandler) ValidateSheet(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	tournamentID := chi.URLParam(r, "id")
	result, err := h.deps.Imports.Validate(r.Context(), actorUserID, tournamentID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *ImportHandler) PreviewImport(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	tournamentID := chi.URLParam(r, "id")
	batch, rows, err := h.deps.Imports.Preview(r.Context(), actorUserID, tournamentID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"batch": batch, "rows": rows})
}

func (h *ImportHandler) ConfirmImport(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	tournamentID := chi.URLParam(r, "id")
	var req confirmImportRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	batch, err := h.deps.Imports.Confirm(r.Context(), actorUserID, tournamentID, req.BatchID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"batch": batch})
}

func (h *ImportHandler) ListImports(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	tournamentID := chi.URLParam(r, "id")
	batches, err := h.deps.Imports.ListBatches(r.Context(), actorUserID, tournamentID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": batches})
}

func (h *ImportHandler) GetImport(w http.ResponseWriter, r *http.Request) {
	actorUserID := mustUserID(r)
	if actorUserID == "" {
		writeError(w, apperror.Unauthorized("missing auth context"))
		return
	}
	batchID := chi.URLParam(r, "batchId")
	batch, rows, err := h.deps.Imports.GetBatch(r.Context(), actorUserID, batchID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"batch": batch, "rows": rows})
}
