package tonic

import (
	"github.com/G-Node/tonic/tonic/worker"
	"testing"
)

func TestTonicFailStart(t *testing.T) {
	if s, _ := NewService(nil, nil); s.Start() == nil {
		s.Stop()
		t.Fatal("Service start succeeded; should have failed")
	}

	if s, _ := NewService(make([]Element, 0), nil); s.Start() == nil {
		s.Stop()
		t.Fatal("Service start succeeded; should have failed")
	}

	if s, _ := NewService(make([]Element, 10), nil); s.Start() == nil {
		s.Stop()
		t.Fatal("Service start succeeded; should have failed")
	}
}

func TestTonicFull(t *testing.T) {
	f := []Element{
		{
			ID:          "el1",
			Name:        "testfield1",
			Label:       "TestField1",
			Description: "Field of tests",
		},
		{
			ID:          "el2",
			Name:        "testfield2",
			Label:       "TestField2",
			Description: "Field of tests, part 2",
		},
	}
	srv, err := NewService(f, testAction)
	if err != nil {
		t.Fatalf("Failed to initialise tonic service: %s", err.Error())
	}
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start tonic service: %s", err.Error())
	}

	srv.Stop()

}

func testAction(values map[string]string, _, _ *worker.Client) ([]string, error) {
	return nil, nil
}
