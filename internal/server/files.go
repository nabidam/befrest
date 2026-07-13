package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/nabidam/befrest/internal/transfer"
)

func (h *webSocketHub) serveFiles(writer http.ResponseWriter, request *http.Request) {
	id, index, ok := transferPath(request.URL.Path)
	if !ok {
		http.NotFound(writer, request)
		return
	}
	if request.Method == http.MethodPost {
		h.uploadFile(writer, request, id, index)
		return
	}
	if request.Method == http.MethodGet {
		h.downloadFile(writer, request, id, index)
		return
	}
	writer.Header().Set("Allow", "GET, POST")
	http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
}

func (h *webSocketHub) uploadFile(writer http.ResponseWriter, request *http.Request, id string, index int) {
	if request.ContentLength < 0 {
		http.Error(writer, "content length required", http.StatusLengthRequired)
		return
	}
	_, file, err := h.transfers.File(id, index)
	if err != nil {
		writeTransferError(writer, err)
		return
	}
	if request.ContentLength != file.Size {
		http.Error(writer, "content length does not match file size", http.StatusBadRequest)
		return
	}
	if err := h.transfers.Upload(id, index, request.Body); err != nil {
		writeTransferError(writer, err)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, _ = writer.Write([]byte(`{"ok":true}`))
}

func (h *webSocketHub) downloadFile(writer http.ResponseWriter, request *http.Request, id string, index int) {
	file, err := h.transfers.Ready(id, index)
	if err != nil {
		writeTransferError(writer, err)
		return
	}
	writer.Header().Set("Content-Type", "application/octet-stream")
	writer.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))
	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", file.Name))
	if _, err := h.transfers.Download(id, index, writer); err != nil {
		return
	}
}

func writeTransferError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, transfer.ErrNotFound):
		http.NotFound(writer, nil)
	case errors.Is(err, transfer.ErrWrongState), errors.Is(err, transfer.ErrNotReceiver):
		http.Error(writer, "transfer is not ready", http.StatusConflict)
	case errors.Is(err, transfer.ErrInvalidFile):
		http.Error(writer, "invalid file", http.StatusBadRequest)
	default:
		http.Error(writer, "transfer failed", http.StatusInternalServerError)
	}
}

func transferPath(path string) (string, int, bool) {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) != 5 || parts[0] != "api" || parts[1] != "transfers" || parts[3] != "files" || parts[2] == "" {
		return "", 0, false
	}
	index, err := strconv.Atoi(parts[4])
	if err != nil || index < 0 {
		return "", 0, false
	}
	return parts[2], index, true
}
