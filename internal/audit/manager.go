package audit

import "github.com/rs/zerolog/log"

// Manager is the Observable that holds a list of Sinks (Observers)
type Manager struct {
	sinks []Sink
}

// NewManager creates a new audit manager with the given sinks
func NewManager(sinks ...Sink) *Manager {
	return &Manager{
		sinks: sinks,
	}
}

// Notify sends the audit event to all registered sinks
func (m *Manager) Notify(event AuditEvent) {
	for _, sink := range m.sinks {
		if err := sink.Send(event); err != nil {
			log.Error().Err(err).Msg("Failed to send audit event to sink")
		}
	}
}
