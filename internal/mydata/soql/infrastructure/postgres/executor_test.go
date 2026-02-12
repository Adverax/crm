package postgres

import (
	"testing"
)

func TestRecordGet(t *testing.T) {
	record := Record{
		Type: "Account",
		Fields: map[string]any{
			"Name":    "Acme Corp",
			"Revenue": 1000000.50,
			"Count":   int64(42),
			"Active":  true,
			"NullVal": nil,
		},
	}

	t.Run("Get existing field", func(t *testing.T) {
		v := record.Get("Name")
		if v != "Acme Corp" {
			t.Errorf("Get(Name) = %v, want Acme Corp", v)
		}
	})

	t.Run("Get non-existing field", func(t *testing.T) {
		v := record.Get("Unknown")
		if v != nil {
			t.Errorf("Get(Unknown) = %v, want nil", v)
		}
	})

	t.Run("GetString", func(t *testing.T) {
		s := record.GetString("Name")
		if s != "Acme Corp" {
			t.Errorf("GetString(Name) = %s, want Acme Corp", s)
		}
	})

	t.Run("GetString on non-string", func(t *testing.T) {
		s := record.GetString("Revenue")
		if s != "" {
			t.Errorf("GetString(Revenue) = %s, want empty string", s)
		}
	})

	t.Run("GetFloat", func(t *testing.T) {
		f := record.GetFloat("Revenue")
		if f != 1000000.50 {
			t.Errorf("GetFloat(Revenue) = %f, want 1000000.50", f)
		}
	})

	t.Run("GetInt", func(t *testing.T) {
		i := record.GetInt("Count")
		if i != 42 {
			t.Errorf("GetInt(Count) = %d, want 42", i)
		}
	})

	t.Run("GetBool", func(t *testing.T) {
		b := record.GetBool("Active")
		if !b {
			t.Errorf("GetBool(Active) = %v, want true", b)
		}
	})

	t.Run("IsNull", func(t *testing.T) {
		if !record.IsNull("NullVal") {
			t.Error("IsNull(NullVal) should be true")
		}
		if record.IsNull("Name") {
			t.Error("IsNull(Name) should be false")
		}
		if !record.IsNull("Unknown") {
			t.Error("IsNull(Unknown) should be true")
		}
	})
}

func TestRecordRelationships(t *testing.T) {
	record := Record{
		Type:   "Account",
		Fields: map[string]any{"Name": "Acme"},
		Relationships: map[string][]Record{
			"Contacts": {
				{Type: "Contact", Fields: map[string]any{"FirstName": "John"}},
				{Type: "Contact", Fields: map[string]any{"FirstName": "Jane"}},
			},
		},
	}

	t.Run("GetRelationship existing", func(t *testing.T) {
		contacts := record.GetRelationship("Contacts")
		if len(contacts) != 2 {
			t.Errorf("GetRelationship(Contacts) length = %d, want 2", len(contacts))
		}
		if contacts[0].GetString("FirstName") != "John" {
			t.Errorf("First contact name = %s, want John", contacts[0].GetString("FirstName"))
		}
	})

	t.Run("GetRelationship non-existing", func(t *testing.T) {
		rel := record.GetRelationship("Unknown")
		if rel != nil {
			t.Errorf("GetRelationship(Unknown) = %v, want nil", rel)
		}
	})
}

func TestQueryResult(t *testing.T) {
	result := &QueryResult{
		Records: []Record{
			{Type: "Account", Fields: map[string]any{"Name": "A"}},
			{Type: "Account", Fields: map[string]any{"Name": "B"}},
		},
		TotalSize: 2,
		Done:      true,
	}

	if len(result.Records) != 2 {
		t.Errorf("Records count = %d, want 2", len(result.Records))
	}

	if result.TotalSize != 2 {
		t.Errorf("TotalSize = %d, want 2", result.TotalSize)
	}

	if !result.Done {
		t.Error("Done should be true")
	}
}

func TestEmptyRecord(t *testing.T) {
	record := Record{}

	if v := record.Get("anything"); v != nil {
		t.Errorf("Get on empty record = %v, want nil", v)
	}

	if s := record.GetString("anything"); s != "" {
		t.Errorf("GetString on empty record = %s, want empty", s)
	}

	if !record.IsNull("anything") {
		t.Error("IsNull on empty record should be true")
	}

	if rel := record.GetRelationship("anything"); rel != nil {
		t.Errorf("GetRelationship on empty record = %v, want nil", rel)
	}
}
