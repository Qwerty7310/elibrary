package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"elibrary/internal/service"
)

type PrintHandler struct {
	Queue *service.PrintQueue
}

func NewPrintHandler(queue *service.PrintQueue) *PrintHandler {
	return &PrintHandler{Queue: queue}
}

type printRequest struct {
	Str1    string `json:"str1"`
	Str2    string `json:"str2"`
	Barcode string `json:"barcode"`
}

func (h *PrintHandler) Send(w http.ResponseWriter, r *http.Request) {
	var req printRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode print request body: %v", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.Str1 = strings.TrimSpace(req.Str1)
	req.Str2 = strings.TrimSpace(req.Str2)
	req.Barcode = strings.TrimSpace(req.Barcode)

	if req.Barcode == "" {
		http.Error(w, "barcode is required", http.StatusBadRequest)
		return
	}

	req.Str1 = service.TransliterateRuToEn(req.Str1)
	req.Str2 = service.TransliterateRuToEn(req.Str2)

	task := service.PrintTask{
		Str1:    req.Str1,
		Str2:    req.Str2,
		Barcode: req.Barcode,
	}

	if err := h.Queue.Send(r.Context(), task); err != nil {
		log.Printf("Failed to enqueue print task: %v", err)
		http.Error(w, "failed to enqueue print task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
