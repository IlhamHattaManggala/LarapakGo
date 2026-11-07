package ngapak

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/IlhamHattaManggala/larapak-go/pkg/rikatlive"
	"github.com/IlhamHattaManggala/larapak-go/pkg/view"
)

func TestNgapakRikatLiveComponent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "rikatlive_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create folder structure
	liveDir := filepath.Join(tempDir, "jeroan", "Urip")
	viewsDir := filepath.Join(tempDir, "sawangan", "live")
	_ = os.MkdirAll(liveDir, 0755)
	_ = os.MkdirAll(viewsDir, 0755)

	// 1. Write Counter.ngpk
	ngpkContent := `
count = 0
jeneng = "Slamet"

gawe tambah(step)
    nek step == kosong ya
        step = 1
    rampung
    count = count + step
rampung

gawe kurang(step)
    nek step == kosong ya
        step = 1
    rampung
    count = count - step
rampung
`
	ngpkFile := filepath.Join(liveDir, "Counter.ngpk")
	if err := os.WriteFile(ngpkFile, []byte(ngpkContent), 0644); err != nil {
		t.Fatalf("failed to write Counter.ngpk: %v", err)
	}

	// 2. Write counter.nview
	nviewContent := `<div class="counter-box">
    <span>Halo, {{ jeneng }}</span>
    <span class="count-val">{{ count }}</span>
    <button r-click="tambah(1)">Plus</button>
</div>`
	nviewFile := filepath.Join(viewsDir, "counter.nview")
	if err := os.WriteFile(nviewFile, []byte(nviewContent), 0644); err != nil {
		t.Fatalf("failed to write counter.nview: %v", err)
	}

	// Setup Engines
	vEngine := view.New(filepath.Join(tempDir, "sawangan"))
	ngpkEngine := NewEngine(tempDir, vEngine)
	RegisterNgapakComponentLoader(tempDir, ngpkEngine, vEngine)

	// 3. Test Initial SSR Render
	html, err := rikatlive.RenderInitial("Counter")
	if err != nil {
		t.Fatalf("RenderInitial failed: %v", err)
	}

	if !strings.Contains(html, `r-component="Counter"`) {
		t.Errorf("expected r-component='Counter', got: %s", html)
	}
	if !strings.Contains(html, "Halo, Slamet") {
		t.Errorf("expected 'Halo, Slamet' in rendered html, got: %s", html)
	}
	if !strings.Contains(html, `class="count-val">0</span>`) {
		t.Errorf("expected count 0 in rendered html, got: %s", html)
	}

	// 4. Test Action Execution (Calling tambah(3))
	// Extract snapshot and checksum
	comp, err := rikatlive.Create("Counter")
	if err != nil {
		t.Fatalf("Create Counter failed: %v", err)
	}
	initState := rikatlive.ExtractState(comp)
	if initState["jeneng"] != "Slamet" || initState["count"] != int64(0) {
		t.Fatalf("unexpected initial state: %v", initState)
	}

	snapJSON, _ := os.ReadFile(ngpkFile) // just non-empty
	_ = snapJSON
	checksum := rikatlive.SignSnapshot([]byte(`{"count":0,"jeneng":"Slamet"}`))

	reqPayload := rikatlive.RequestPayload{
		Component: "Counter",
		ID:        "r-test-1",
		Snapshot:  map[string]any{"count": int64(0), "jeneng": "Slamet"},
		Checksum:  checksum,
		Action:    "tambah",
		Args:      []any{int64(3)},
	}

	res, err := rikatlive.ProcessRequest(reqPayload)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if res.Snapshot["count"] != int64(3) {
		t.Errorf("expected count 3, got: %v", res.Snapshot["count"])
	}
	if !strings.Contains(res.HTML, `class="count-val">3</span>`) {
		t.Errorf("expected count 3 in returned html, got: %s", res.HTML)
	}

	// 5. Test Two-Way Model Update (Updating jeneng)
	checksum2 := rikatlive.SignSnapshot([]byte(`{"count":3,"jeneng":"Slamet"}`))
	reqPayload2 := rikatlive.RequestPayload{
		Component: "Counter",
		ID:        "r-test-1",
		Snapshot:  map[string]any{"count": int64(3), "jeneng": "Slamet"},
		Checksum:  checksum2,
		Updates:   map[string]any{"jeneng": "Bambang"},
	}

	res2, err := rikatlive.ProcessRequest(reqPayload2)
	if err != nil {
		t.Fatalf("ProcessRequest with update failed: %v", err)
	}

	if res2.Snapshot["jeneng"] != "Bambang" {
		t.Errorf("expected jeneng 'Bambang', got: %v", res2.Snapshot["jeneng"])
	}
	if !strings.Contains(res2.HTML, "Halo, Bambang") {
		t.Errorf("expected 'Halo, Bambang' in returned html, got: %s", res2.HTML)
	}
}

