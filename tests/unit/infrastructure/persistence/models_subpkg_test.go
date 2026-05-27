package persistence_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence/models"
)

func TestModels_BeforeCreate(t *testing.T) {
	t.Run("BaseModel with nil ID", func(t *testing.T) {
		b := &models.BaseModel{}
		if err := b.BeforeCreate(nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if b.ID == uuid.Nil {
			t.Error("expected non-nil UUID")
		}
	})

	t.Run("BaseModel with existing ID", func(t *testing.T) {
		existing := uuid.New()
		b := &models.BaseModel{ID: existing}
		if err := b.BeforeCreate(nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if b.ID != existing {
			t.Error("expected existing ID to remain")
		}
	})

	t.Run("Session BeforeCreate", func(t *testing.T) {
		s := &models.Session{}
		if err := s.BeforeCreate(nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.ID == uuid.Nil {
			t.Error("expected non-nil UUID")
		}
	})

	t.Run("Conversation BeforeCreate", func(t *testing.T) {
		c := &models.Conversation{}
		if err := c.BeforeCreate(nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ID == uuid.Nil {
			t.Error("expected non-nil UUID")
		}
	})

	t.Run("Message BeforeCreate", func(t *testing.T) {
		m := &models.Message{}
		if err := m.BeforeCreate(nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m.ID == uuid.Nil {
			t.Error("expected non-nil UUID")
		}
	})

	t.Run("Tool BeforeCreate", func(t *testing.T) {
		tool := &models.Tool{}
		if err := tool.BeforeCreate(nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tool.ID == uuid.Nil {
			t.Error("expected non-nil UUID")
		}
	})

	t.Run("Resource BeforeCreate", func(t *testing.T) {
		r := &models.Resource{}
		if err := r.BeforeCreate(nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.ID == uuid.Nil {
			t.Error("expected non-nil UUID")
		}
	})

	t.Run("Prompt BeforeCreate", func(t *testing.T) {
		p := &models.Prompt{}
		if err := p.BeforeCreate(nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.ID == uuid.Nil {
			t.Error("expected non-nil UUID")
		}
	})

	t.Run("ResourceSubscription BeforeCreate", func(t *testing.T) {
		rs := &models.ResourceSubscription{}
		if err := rs.BeforeCreate(nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rs.ID == uuid.Nil {
			t.Error("expected non-nil UUID")
		}
	})

	t.Run("ToolExecution BeforeCreate", func(t *testing.T) {
		te := &models.ToolExecution{}
		if err := te.BeforeCreate(nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if te.ID == uuid.Nil {
			t.Error("expected non-nil UUID")
		}
	})

	t.Run("APIKey BeforeCreate", func(t *testing.T) {
		ak := &models.APIKey{}
		if err := ak.BeforeCreate(nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ak.ID == uuid.Nil {
			t.Error("expected non-nil UUID")
		}
	})
}

func TestModels_TableNames(t *testing.T) {
	rs := models.ResourceSubscription{}
	if rs.TableName() != "resource_subscriptions" {
		t.Error("unexpected table name")
	}
	te := models.ToolExecution{}
	if te.TableName() != "tool_executions" {
		t.Error("unexpected table name")
	}
	ak := models.APIKey{}
	if ak.TableName() != "api_keys" {
		t.Error("unexpected table name")
	}
	sm := models.SchemaMigration{}
	if sm.TableName() != "schema_migrations" {
		t.Error("unexpected table name")
	}
}

func TestModels_JSONB_ScanValue(t *testing.T) {
	t.Run("JSONB Scan with bytes", func(t *testing.T) {
		var j models.JSONB
		if err := j.Scan([]byte(`{"key":"value"}`)); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if j["key"] != "value" {
			t.Errorf("expected value, got %v", j["key"])
		}
	})

	t.Run("JSONB Scan with nil", func(t *testing.T) {
		var j models.JSONB
		if err := j.Scan(nil); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if j != nil {
			t.Errorf("expected nil, got %v", j)
		}
	})

	t.Run("JSONB Scan with wrong type", func(t *testing.T) {
		var j models.JSONB
		err := j.Scan("not bytes")
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, errors.New("type assertion to []byte failed")) {
			t.Logf("wrong type error: %v", err)
		}
	})

	t.Run("JSONB Value with data", func(t *testing.T) {
		j := models.JSONB{"key": "value"}
		v, err := j.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}
		if v == nil {
			t.Fatal("expected non-nil value")
		}
	})

	t.Run("JSONB Value nil", func(t *testing.T) {
		var j models.JSONB
		v, err := j.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}
		if v != nil {
			t.Errorf("expected nil, got %v", v)
		}
	})
}

func TestModels_JSONBArray_ScanValue(t *testing.T) {
	t.Run("JSONBArray Scan with bytes", func(t *testing.T) {
		var j models.JSONBArray
		if err := j.Scan([]byte(`[1,2,3]`)); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if len(j) != 3 {
			t.Errorf("expected 3 elements, got %d", len(j))
		}
	})

	t.Run("JSONBArray Scan with nil", func(t *testing.T) {
		var j models.JSONBArray
		if err := j.Scan(nil); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if j != nil {
			t.Errorf("expected nil, got %v", j)
		}
	})

	t.Run("JSONBArray Scan with wrong type", func(t *testing.T) {
		var j models.JSONBArray
		err := j.Scan(123)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("JSONBArray Value with data", func(t *testing.T) {
		j := models.JSONBArray{"a", "b"}
		v, err := j.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}
		if v == nil {
			t.Fatal("expected non-nil value")
		}
	})

	t.Run("JSONBArray Value nil", func(t *testing.T) {
		var j models.JSONBArray
		v, err := j.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}
		if v != nil {
			t.Errorf("expected nil, got %v", v)
		}
	})
}

func TestModels_StringArray_ScanValue(t *testing.T) {
	t.Run("StringArray Scan with bytes", func(t *testing.T) {
		var s models.StringArray
		if err := s.Scan([]byte(`["a","b","c"]`)); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if len(s) != 3 {
			t.Errorf("expected 3 elements, got %d", len(s))
		}
		if s[0] != "a" {
			t.Errorf("expected a, got %s", s[0])
		}
	})

	t.Run("StringArray Scan with nil", func(t *testing.T) {
		var s models.StringArray
		if err := s.Scan(nil); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if s != nil {
			t.Errorf("expected nil, got %v", s)
		}
	})

	t.Run("StringArray Scan with wrong type", func(t *testing.T) {
		var s models.StringArray
		err := s.Scan(42)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("StringArray Value with data", func(t *testing.T) {
		s := models.StringArray{"x", "y"}
		v, err := s.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}
		if v == nil {
			t.Fatal("expected non-nil value")
		}
	})

	t.Run("StringArray Value nil", func(t *testing.T) {
		var s models.StringArray
		v, err := s.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}
		if v != nil {
			t.Errorf("expected nil, got %v", v)
		}
	})
}

func TestModels_AllModels(t *testing.T) {
	all := models.AllModels()
	if len(all) != 10 {
		t.Errorf("expected 10 models, got %d", len(all))
	}
}
