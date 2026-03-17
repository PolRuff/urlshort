package handler

import (
	"github.com/PolRuff/urlshort/internal/audit"
	"github.com/PolRuff/urlshort/internal/repository"
)

func newTestHandler() *Handler {
	repo := repository.NewMemoryRepository()
	var auditSinks []audit.Sink
	auditManager := audit.NewManager(auditSinks...)
	return New(repo, "http://localhost:8080", "very-strong-key", auditManager, nil)
}
