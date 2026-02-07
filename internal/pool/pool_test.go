package pool

import (
	"sync/atomic"
	"testing"
)

// Следующий код не должен компилироваться,
// если раскомментировать:
//
// type NoReset struct{}
//
// func TestCompileTimeConstraint(t *testing.T) {
//     _ = New(func() *NoReset {
//         return &NoReset{}
//     })
// }

func TestGetReturnsObject(t *testing.T) {
	p := New(func() *TestStruct {
		return &TestStruct{
			Tags:     make([]string, 0, 4),
			Settings: make(map[string]bool),
		}
	})

	obj := p.Get()
	if obj == nil {
		t.Fatal("Get returned nil")
	}
}

func TestPutResetsState(t *testing.T) {
	p := New(func() *TestStruct {
		return &TestStruct{
			Tags:     make([]string, 0, 4),
			Settings: make(map[string]bool),
		}
	})

	obj := p.Get()

	obj.Value = 100
	obj.Name = "test"
	obj.Tags = append(obj.Tags, "a", "b")
	obj.Settings["debug"] = true
	obj.Nested = &TestStruct{
		Value: 10,
		Name:  "nested",
		Tags:  []string{"x"},
		Settings: map[string]bool{
			"enabled": true,
		},
	}

	p.Put(obj)

	if obj.Value != 0 {
		t.Errorf("Value not reset, got %d", obj.Value)
	}

	if obj.Name != "" {
		t.Errorf("Name not reset, got %q", obj.Name)
	}

	if len(obj.Tags) != 0 {
		t.Errorf("Tags not reset, len = %d", len(obj.Tags))
	}

	if len(obj.Settings) != 0 {
		t.Errorf("Settings not reset, len = %d", len(obj.Settings))
	}

	if obj.Nested == nil {
		t.Fatal("Nested struct is nil after Put")
	}

	if obj.Nested.Value != 0 {
		t.Errorf("Nested.Value not reset, got %d", obj.Nested.Value)
	}

	if obj.Nested.Name != "" {
		t.Errorf("Nested.Name not reset, got %q", obj.Nested.Name)
	}

	if len(obj.Nested.Tags) != 0 {
		t.Errorf("Nested.Tags not reset, len = %d", len(obj.Nested.Tags))
	}

	if len(obj.Nested.Settings) != 0 {
		t.Errorf("Nested.Settings not reset, len = %d", len(obj.Nested.Settings))
	}
}

func TestObjectReuse(t *testing.T) {
	var created int32

	p := New(func() *TestStruct {
		atomic.AddInt32(&created, 1)
		return &TestStruct{
			Tags:     make([]string, 0, 4),
			Settings: make(map[string]bool),
		}
	})

	obj1 := p.Get()
	p.Put(obj1)

	obj2 := p.Get()

	if obj1 != obj2 {
		t.Error("expected object to be reused from pool")
	}

	if atomic.LoadInt32(&created) != 1 {
		t.Errorf("expected 1 object to be created, got %d", created)
	}
}

func TestResetOnNilDoesNotPanic(t *testing.T) {
	var rs *TestStruct

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Reset panicked on nil receiver: %v", r)
		}
	}()

	rs.Reset()
}
