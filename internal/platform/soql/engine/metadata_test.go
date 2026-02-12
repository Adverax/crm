package engine

import (
	"context"
	"testing"
)

func TestObjectMetaGetField(t *testing.T) {
	obj := &ObjectMeta{
		Name:      "Account",
		TableName: "accounts",
		Fields: map[string]*FieldMeta{
			"Name":     {Name: "Name", Column: "name", Type: FieldTypeString},
			"Industry": {Name: "Industry", Column: "industry", Type: FieldTypeString},
		},
	}

	t.Run("existing field", func(t *testing.T) {
		field := obj.GetField("Name")
		if field == nil {
			t.Error("expected field, got nil")
		} else if field.Name != "Name" {
			t.Errorf("field.Name = %s, want Name", field.Name)
		}
	})

	t.Run("non-existing field", func(t *testing.T) {
		field := obj.GetField("Unknown")
		if field != nil {
			t.Error("expected nil for unknown field")
		}
	})

	t.Run("nil object", func(t *testing.T) {
		var nilObj *ObjectMeta
		field := nilObj.GetField("Name")
		if field != nil {
			t.Error("expected nil for nil object")
		}
	})

	t.Run("nil fields map", func(t *testing.T) {
		emptyObj := &ObjectMeta{Name: "Empty"}
		field := emptyObj.GetField("Name")
		if field != nil {
			t.Error("expected nil for nil fields map")
		}
	})
}

func TestObjectMetaGetFieldByColumn(t *testing.T) {
	obj := &ObjectMeta{
		Name:      "Contact",
		TableName: "contacts",
		Fields: map[string]*FieldMeta{
			"FirstName": {Name: "FirstName", Column: "first_name", Type: FieldTypeString},
			"LastName":  {Name: "LastName", Column: "last_name", Type: FieldTypeString},
			"Email":     {Name: "Email", Column: "email", Type: FieldTypeString},
		},
	}

	t.Run("existing column", func(t *testing.T) {
		field := obj.GetFieldByColumn("first_name")
		if field == nil {
			t.Error("expected field, got nil")
		} else if field.Name != "FirstName" {
			t.Errorf("field.Name = %s, want FirstName", field.Name)
		}
	})

	t.Run("non-existing column", func(t *testing.T) {
		field := obj.GetFieldByColumn("unknown_column")
		if field != nil {
			t.Error("expected nil for unknown column")
		}
	})

	t.Run("nil object", func(t *testing.T) {
		var nilObj *ObjectMeta
		field := nilObj.GetFieldByColumn("first_name")
		if field != nil {
			t.Error("expected nil for nil object")
		}
	})

	t.Run("nil fields map", func(t *testing.T) {
		emptyObj := &ObjectMeta{Name: "Empty"}
		field := emptyObj.GetFieldByColumn("first_name")
		if field != nil {
			t.Error("expected nil for nil fields map")
		}
	})
}

func TestObjectMetaGetLookup(t *testing.T) {
	obj := &ObjectMeta{
		Name:      "Contact",
		TableName: "contacts",
		Lookups: map[string]*LookupMeta{
			"Account": {Name: "Account", Field: "account_id", TargetObject: "Account", TargetField: "id"},
		},
	}

	t.Run("existing lookup", func(t *testing.T) {
		lookup := obj.GetLookup("Account")
		if lookup == nil {
			t.Error("expected lookup, got nil")
		} else if lookup.TargetObject != "Account" {
			t.Errorf("lookup.TargetObject = %s, want Account", lookup.TargetObject)
		}
	})

	t.Run("non-existing lookup", func(t *testing.T) {
		lookup := obj.GetLookup("Unknown")
		if lookup != nil {
			t.Error("expected nil for unknown lookup")
		}
	})

	t.Run("nil object", func(t *testing.T) {
		var nilObj *ObjectMeta
		lookup := nilObj.GetLookup("Account")
		if lookup != nil {
			t.Error("expected nil for nil object")
		}
	})

	t.Run("nil lookups map", func(t *testing.T) {
		emptyObj := &ObjectMeta{Name: "Empty"}
		lookup := emptyObj.GetLookup("Account")
		if lookup != nil {
			t.Error("expected nil for nil lookups map")
		}
	})
}

func TestObjectMetaGetRelationship(t *testing.T) {
	obj := &ObjectMeta{
		Name:      "Account",
		TableName: "accounts",
		Relationships: map[string]*RelationshipMeta{
			"Contacts": {Name: "Contacts", ChildObject: "Contact", ChildField: "account_id", ParentField: "id"},
		},
	}

	t.Run("existing relationship", func(t *testing.T) {
		rel := obj.GetRelationship("Contacts")
		if rel == nil {
			t.Error("expected relationship, got nil")
		} else if rel.ChildObject != "Contact" {
			t.Errorf("rel.ChildObject = %s, want Contact", rel.ChildObject)
		}
	})

	t.Run("non-existing relationship", func(t *testing.T) {
		rel := obj.GetRelationship("Unknown")
		if rel != nil {
			t.Error("expected nil for unknown relationship")
		}
	})

	t.Run("nil object", func(t *testing.T) {
		var nilObj *ObjectMeta
		rel := nilObj.GetRelationship("Contacts")
		if rel != nil {
			t.Error("expected nil for nil object")
		}
	})

	t.Run("nil relationships map", func(t *testing.T) {
		emptyObj := &ObjectMeta{Name: "Empty"}
		rel := emptyObj.GetRelationship("Contacts")
		if rel != nil {
			t.Error("expected nil for nil relationships map")
		}
	})
}

func TestStaticMetadataProvider(t *testing.T) {
	ctx := context.Background()

	objects := map[string]*ObjectMeta{
		"Account": {
			Name:      "Account",
			TableName: "accounts",
			Fields: map[string]*FieldMeta{
				"Name": {Name: "Name", Column: "name", Type: FieldTypeString},
			},
		},
		"Contact": {
			Name:      "Contact",
			TableName: "contacts",
			Fields: map[string]*FieldMeta{
				"Email": {Name: "Email", Column: "email", Type: FieldTypeString},
			},
		},
	}

	provider := NewStaticMetadataProvider(objects)

	t.Run("GetObject existing", func(t *testing.T) {
		obj, err := provider.GetObject(ctx, "Account")
		if err != nil {
			t.Fatalf("GetObject() error = %v", err)
		}
		if obj == nil {
			t.Error("expected object, got nil")
		} else if obj.Name != "Account" {
			t.Errorf("obj.Name = %s, want Account", obj.Name)
		}
	})

	t.Run("GetObject non-existing", func(t *testing.T) {
		obj, err := provider.GetObject(ctx, "Unknown")
		if err != nil {
			t.Fatalf("GetObject() error = %v", err)
		}
		if obj != nil {
			t.Error("expected nil for unknown object")
		}
	})

	t.Run("ListObjects", func(t *testing.T) {
		names, err := provider.ListObjects(ctx)
		if err != nil {
			t.Fatalf("ListObjects() error = %v", err)
		}
		if len(names) != 2 {
			t.Errorf("len(names) = %d, want 2", len(names))
		}
		// Check that both objects are in the list
		hasAccount := false
		hasContact := false
		for _, name := range names {
			if name == "Account" {
				hasAccount = true
			}
			if name == "Contact" {
				hasContact = true
			}
		}
		if !hasAccount {
			t.Error("expected Account in list")
		}
		if !hasContact {
			t.Error("expected Contact in list")
		}
	})
}

func TestStaticMetadataProviderNil(t *testing.T) {
	ctx := context.Background()
	provider := NewStaticMetadataProvider(nil)

	t.Run("GetObject with nil map", func(t *testing.T) {
		obj, err := provider.GetObject(ctx, "Account")
		if err != nil {
			t.Fatalf("GetObject() error = %v", err)
		}
		if obj != nil {
			t.Error("expected nil for nil map")
		}
	})

	t.Run("ListObjects with nil map", func(t *testing.T) {
		names, err := provider.ListObjects(ctx)
		if err != nil {
			t.Fatalf("ListObjects() error = %v", err)
		}
		if names != nil {
			t.Error("expected nil for nil map")
		}
	})
}

func TestObjectMetaBuilder(t *testing.T) {
	obj := NewObjectMeta("Account", "", "accounts").
		Field("Id", "id", FieldTypeID).
		Field("Name", "name", FieldTypeString).
		Field("Amount", "amount", FieldTypeFloat).
		Lookup("Owner", "owner_id", "User", "id").
		Relationship("Contacts", "Contact", "account_id", "id").
		Build()

	t.Run("basic properties", func(t *testing.T) {
		if obj.Name != "Account" {
			t.Errorf("Name = %s, want Account", obj.Name)
		}
		if obj.TableName != "accounts" {
			t.Errorf("TableName = %s, want accounts", obj.TableName)
		}
	})

	t.Run("fields", func(t *testing.T) {
		if len(obj.Fields) != 3 {
			t.Errorf("len(Fields) = %d, want 3", len(obj.Fields))
		}

		id := obj.Fields["Id"]
		if id == nil {
			t.Error("expected Id field")
		} else {
			if id.Column != "id" {
				t.Errorf("Id.Column = %s, want id", id.Column)
			}
			if id.Type != FieldTypeID {
				t.Errorf("Id.Type = %v, want FieldTypeID", id.Type)
			}
		}

		name := obj.Fields["Name"]
		if name == nil {
			t.Error("expected Name field")
		} else {
			if name.Column != "name" {
				t.Errorf("Name.Column = %s, want name", name.Column)
			}
			if name.Type != FieldTypeString {
				t.Errorf("Name.Type = %v, want FieldTypeString", name.Type)
			}
		}
	})

	t.Run("field defaults", func(t *testing.T) {
		name := obj.Fields["Name"]
		if !name.Filterable {
			t.Error("expected Filterable = true")
		}
		if !name.Sortable {
			t.Error("expected Sortable = true")
		}
		if !name.Groupable {
			t.Error("expected Groupable = true")
		}
		if !name.Nullable {
			t.Error("expected Nullable = true")
		}
	})

	t.Run("numeric field aggregatable", func(t *testing.T) {
		amount := obj.Fields["Amount"]
		if !amount.Aggregatable {
			t.Error("expected numeric field to be Aggregatable")
		}

		name := obj.Fields["Name"]
		if name.Aggregatable {
			t.Error("expected string field to not be Aggregatable")
		}
	})

	t.Run("lookup", func(t *testing.T) {
		if len(obj.Lookups) != 1 {
			t.Errorf("len(Lookups) = %d, want 1", len(obj.Lookups))
		}

		owner := obj.Lookups["Owner"]
		if owner == nil {
			t.Error("expected Owner lookup")
		} else {
			if owner.Field != "owner_id" {
				t.Errorf("Owner.Field = %s, want owner_id", owner.Field)
			}
			if owner.TargetObject != "User" {
				t.Errorf("Owner.TargetObject = %s, want User", owner.TargetObject)
			}
			if owner.TargetField != "id" {
				t.Errorf("Owner.TargetField = %s, want id", owner.TargetField)
			}
		}
	})

	t.Run("relationship", func(t *testing.T) {
		if len(obj.Relationships) != 1 {
			t.Errorf("len(Relationships) = %d, want 1", len(obj.Relationships))
		}

		contacts := obj.Relationships["Contacts"]
		if contacts == nil {
			t.Error("expected Contacts relationship")
		} else {
			if contacts.ChildObject != "Contact" {
				t.Errorf("Contacts.ChildObject = %s, want Contact", contacts.ChildObject)
			}
			if contacts.ChildField != "account_id" {
				t.Errorf("Contacts.ChildField = %s, want account_id", contacts.ChildField)
			}
			if contacts.ParentField != "id" {
				t.Errorf("Contacts.ParentField = %s, want id", contacts.ParentField)
			}
		}
	})
}

func TestObjectMetaBuilderFieldFull(t *testing.T) {
	customField := &FieldMeta{
		Name:         "SecretField",
		Column:       "secret_column",
		Type:         FieldTypeString,
		Nullable:     false,
		Filterable:   false,
		Sortable:     false,
		Groupable:    false,
		Aggregatable: false,
	}

	obj := NewObjectMeta("Account", "", "accounts").
		FieldFull(customField).
		Build()

	field := obj.Fields["SecretField"]
	if field == nil {
		t.Fatal("expected SecretField")
	}

	if field.Nullable {
		t.Error("expected Nullable = false")
	}
	if field.Filterable {
		t.Error("expected Filterable = false")
	}
	if field.Sortable {
		t.Error("expected Sortable = false")
	}
	if field.Groupable {
		t.Error("expected Groupable = false")
	}
	if field.Aggregatable {
		t.Error("expected Aggregatable = false")
	}
}

func TestFieldMetaTypes(t *testing.T) {
	// Test that all field types can be assigned
	types := []FieldType{
		FieldTypeID,
		FieldTypeString,
		FieldTypeInteger,
		FieldTypeFloat,
		FieldTypeBoolean,
		FieldTypeDate,
		FieldTypeDateTime,
	}

	for _, ft := range types {
		obj := NewObjectMeta("Test", "", "test").
			Field("TestField", "test_field", ft).
			Build()

		field := obj.Fields["TestField"]
		if field.Type != ft {
			t.Errorf("field.Type = %v, want %v", field.Type, ft)
		}
	}
}

func TestLookupMetaFields(t *testing.T) {
	lookup := &LookupMeta{
		Name:         "Account",
		Field:        "account_id",
		TargetObject: "Account",
		TargetField:  "id",
	}

	if lookup.Name != "Account" {
		t.Errorf("Name = %s, want Account", lookup.Name)
	}
	if lookup.Field != "account_id" {
		t.Errorf("Field = %s, want account_id", lookup.Field)
	}
	if lookup.TargetObject != "Account" {
		t.Errorf("TargetObject = %s, want Account", lookup.TargetObject)
	}
	if lookup.TargetField != "id" {
		t.Errorf("TargetField = %s, want id", lookup.TargetField)
	}
}

func TestRelationshipMetaFields(t *testing.T) {
	rel := &RelationshipMeta{
		Name:        "Contacts",
		ChildObject: "Contact",
		ChildField:  "account_id",
		ParentField: "id",
	}

	if rel.Name != "Contacts" {
		t.Errorf("Name = %s, want Contacts", rel.Name)
	}
	if rel.ChildObject != "Contact" {
		t.Errorf("ChildObject = %s, want Contact", rel.ChildObject)
	}
	if rel.ChildField != "account_id" {
		t.Errorf("ChildField = %s, want account_id", rel.ChildField)
	}
	if rel.ParentField != "id" {
		t.Errorf("ParentField = %s, want id", rel.ParentField)
	}
}

func TestMetadataProviderInterfaceCompliance(t *testing.T) {
	// Compile-time interface compliance check
	var _ MetadataProvider = &StaticMetadataProvider{}
}
