package audit

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
		// Игнорируем ошибку от каждого отдельного sink'а,
		// чтобы сбой одного приёмника не влиял на остальные.
		_ = sink.Send(event)
	}
}
