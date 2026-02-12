package pool

// TestStruct is a test structure that implements the Resetter interface.
// generate:reset
type TestStruct struct {
	Value    int
	Name     string
	Tags     []string
	Settings map[string]bool
	Nested   *TestStruct
}
