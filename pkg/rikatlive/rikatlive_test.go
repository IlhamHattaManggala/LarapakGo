package rikatlive

import (
	"fmt"
	"strings"
	"testing"
)

type CounterComponent struct {
	Count int    `json:"count"`
	Name  string `json:"name"`
}

func (c *CounterComponent) Increment() {
	c.Count++
}

func (c *CounterComponent) Decrement() {
	c.Count--
}

func (c *CounterComponent) Add(by int) {
	c.Count += by
}

func (c *CounterComponent) Render() string {
	return fmt.Sprintf("<div class='counter'><p>Halo %s, Itungan: %d</p><button r-click='Increment'>+</button></div>", c.Name, c.Count)
}

func TestRikatLiveLifecycle(t *testing.T) {
	Register("Counter", func() Component {
		return &CounterComponent{Name: "Tegal"}
	})

	// 1. Initial Server-Side Render
	htmlOut, err := RenderInitial("Counter", map[string]any{"count": 5})
	if err != nil {
		t.Fatalf("render initial failed: %v", err)
	}

	if !strings.Contains(htmlOut, `r-component="Counter"`) {
		t.Fatalf("missing r-component attribute: %s", htmlOut)
	}
	if !strings.Contains(htmlOut, `Itungan: 5`) {
		t.Fatalf("expected initial count 5: %s", htmlOut)
	}

	// 2. Action Execution via ProcessRequest (Increment)
	comp, _ := Create("Counter")
	_ = RestoreState(comp, map[string]any{"Count": 5, "Name": "Tegal"})
	initialState := ExtractState(comp)

	// Generate valid checksum for initial state
	reqPayload := RequestPayload{
		Component: "Counter",
		ID:        "r-test1",
		Snapshot:  initialState,
		Checksum:  SignSnapshot([]byte(`{"Count":5,"Name":"Tegal"}`)),
		Action:    "Increment",
	}

	res, err := ProcessRequest(reqPayload)
	if err != nil {
		t.Fatalf("process request failed: %v", err)
	}

	if !strings.Contains(res.HTML, `Itungan: 6`) {
		t.Fatalf("expected count 6 after Increment: %s", res.HTML)
	}

	// 3. Action with argument (Add by 10)
	reqPayload2 := RequestPayload{
		Component: "Counter",
		ID:        "r-test1",
		Snapshot:  res.Snapshot,
		Checksum:  res.Checksum,
		Action:    "Add",
		Args:      []any{10},
	}

	res2, err := ProcessRequest(reqPayload2)
	if err != nil {
		t.Fatalf("action with args failed: %v", err)
	}
	if !strings.Contains(res2.HTML, `Itungan: 16`) {
		t.Fatalf("expected count 16 after Add(10): %s", res2.HTML)
	}

	// 4. Input update (Change Name)
	reqPayload3 := RequestPayload{
		Component: "Counter",
		ID:        "r-test1",
		Snapshot:  res2.Snapshot,
		Checksum:  res2.Checksum,
		Updates:   map[string]any{"Name": "Ngapak"},
	}

	res3, err := ProcessRequest(reqPayload3)
	if err != nil {
		t.Fatalf("update name failed: %v", err)
	}
	if !strings.Contains(res3.HTML, `Halo Ngapak, Itungan: 16`) {
		t.Fatalf("expected updated name: %s", res3.HTML)
	}

	// 5. Anti-Tamper Security Test (Tampered state rejected)
	tamperedPayload := RequestPayload{
		Component: "Counter",
		ID:        "r-test1",
		Snapshot:  map[string]any{"Count": 999999, "Name": "Hacker"},
		Checksum:  res3.Checksum, // Mismatched checksum
		Action:    "Increment",
	}

	_, err = ProcessRequest(tamperedPayload)
	if err == nil || !strings.Contains(err.Error(), "tampered") {
		t.Fatalf("expected tampering error, got: %v", err)
	}
}

