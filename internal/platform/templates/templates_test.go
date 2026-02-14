package templates

import (
	"testing"

	"github.com/adverax/crm/internal/platform/metadata"
)

func TestBuildRegistry(t *testing.T) {
	t.Parallel()

	registry := BuildRegistry()
	templates := registry.List()

	if len(templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(templates))
	}

	tests := []struct {
		name string
		id   string
	}{
		{name: "sales_crm is registered", id: "sales_crm"},
		{name: "recruiting is registered", id: "recruiting"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, ok := registry.Get(tt.id)
			if !ok {
				t.Errorf("template %q not found in registry", tt.id)
			}
		})
	}
}

func TestRegistry_Get_NotFound(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	_, ok := registry.Get("nonexistent")
	if ok {
		t.Error("expected Get to return false for nonexistent template")
	}
}

func TestRegistry_Register_Panic_OnDuplicate(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	registry.Register(Template{ID: "test"})

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate registration")
		}
	}()
	registry.Register(Template{ID: "test"})
}

func TestTemplateStructure(t *testing.T) {
	t.Parallel()

	registry := BuildRegistry()

	for _, tmpl := range registry.List() {
		t.Run(tmpl.ID, func(t *testing.T) {
			t.Parallel()
			validateTemplate(t, tmpl)
		})
	}
}

func validateTemplate(t *testing.T, tmpl Template) {
	t.Helper()

	if tmpl.ID == "" {
		t.Error("template ID is empty")
	}
	if tmpl.Label == "" {
		t.Error("template Label is empty")
	}
	if len(tmpl.Objects) == 0 {
		t.Error("template has no objects")
	}

	// Collect object API names for reference validation.
	objectNames := make(map[string]bool, len(tmpl.Objects))
	for _, obj := range tmpl.Objects {
		if obj.APIName == "" {
			t.Error("object has empty APIName")
		}
		if obj.Label == "" {
			t.Errorf("object %s has empty Label", obj.APIName)
		}
		if obj.PluralLabel == "" {
			t.Errorf("object %s has empty PluralLabel", obj.APIName)
		}
		if objectNames[obj.APIName] {
			t.Errorf("duplicate object APIName: %s", obj.APIName)
		}
		objectNames[obj.APIName] = true
	}

	// Validate fields.
	fieldKeys := make(map[string]bool)
	for _, f := range tmpl.Fields {
		key := f.ObjectAPIName + "." + f.APIName
		if fieldKeys[key] {
			t.Errorf("duplicate field: %s", key)
		}
		fieldKeys[key] = true

		if f.APIName == "" {
			t.Errorf("field in object %s has empty APIName", f.ObjectAPIName)
		}
		if f.Label == "" {
			t.Errorf("field %s has empty Label", key)
		}
		if !objectNames[f.ObjectAPIName] {
			t.Errorf("field %s references unknown object %s", key, f.ObjectAPIName)
		}

		// Reference fields must point to a known object.
		if f.FieldType == metadata.FieldTypeReference {
			if f.ReferencedObjectAPIName == "" {
				t.Errorf("reference field %s has no ReferencedObjectAPIName", key)
			} else if !objectNames[f.ReferencedObjectAPIName] {
				t.Errorf("reference field %s points to unknown object %s", key, f.ReferencedObjectAPIName)
			}
			if f.Config.OnDelete == nil {
				t.Errorf("reference field %s has no OnDelete config", key)
			}
			if f.Config.RelationshipName == nil {
				t.Errorf("reference field %s has no RelationshipName config", key)
			}
		}
	}
}

func TestSalesCRM_FieldCount(t *testing.T) {
	t.Parallel()

	tmpl := SalesCRM()
	if len(tmpl.Objects) != 4 {
		t.Errorf("expected 4 objects, got %d", len(tmpl.Objects))
	}

	fieldCounts := make(map[string]int)
	for _, f := range tmpl.Fields {
		fieldCounts[f.ObjectAPIName]++
	}

	tests := []struct {
		name     string
		object   string
		expected int
	}{
		{name: "account fields", object: "account", expected: 9},
		{name: "contact fields", object: "contact", expected: 9},
		{name: "opportunity fields", object: "opportunity", expected: 9},
		{name: "task fields", object: "task", expected: 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if fieldCounts[tt.object] != tt.expected {
				t.Errorf("expected %d fields for %s, got %d", tt.expected, tt.object, fieldCounts[tt.object])
			}
		})
	}
}

func TestRecruiting_FieldCount(t *testing.T) {
	t.Parallel()

	tmpl := Recruiting()
	if len(tmpl.Objects) != 4 {
		t.Errorf("expected 4 objects, got %d", len(tmpl.Objects))
	}

	fieldCounts := make(map[string]int)
	for _, f := range tmpl.Fields {
		fieldCounts[f.ObjectAPIName]++
	}

	tests := []struct {
		name     string
		object   string
		expected int
	}{
		{name: "position fields", object: "position", expected: 8},
		{name: "candidate fields", object: "candidate", expected: 9},
		{name: "application fields", object: "application", expected: 6},
		{name: "interview fields", object: "interview", expected: 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if fieldCounts[tt.object] != tt.expected {
				t.Errorf("expected %d fields for %s, got %d", tt.expected, tt.object, fieldCounts[tt.object])
			}
		})
	}
}
