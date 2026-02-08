package metadata

import "fmt"

// TypeSpec describes valid configuration requirements for a (type, subtype) pair.
type TypeSpec struct {
	Type           FieldType
	Subtype        *FieldSubtype
	RequiredConfig []string
	OptionalConfig []string
}

var typeRegistry map[string]TypeSpec

func init() {
	typeRegistry = make(map[string]TypeSpec)
	register := func(ft FieldType, fst *FieldSubtype, required, optional []string) {
		key := typeSubtypeKey(ft, fst)
		typeRegistry[key] = TypeSpec{
			Type:           ft,
			Subtype:        fst,
			RequiredConfig: required,
			OptionalConfig: optional,
		}
	}

	sub := func(s FieldSubtype) *FieldSubtype { return &s }

	// text subtypes
	register(FieldTypeText, sub(SubtypePlain), []string{"max_length"}, []string{"default_value"})
	register(FieldTypeText, sub(SubtypeArea), nil, []string{"default_value"})
	register(FieldTypeText, sub(SubtypeRich), nil, []string{"default_value"})
	register(FieldTypeText, sub(SubtypeEmail), nil, []string{"default_value"})
	register(FieldTypeText, sub(SubtypePhone), nil, []string{"default_value"})
	register(FieldTypeText, sub(SubtypeURL), nil, []string{"default_value"})

	// number subtypes
	register(FieldTypeNumber, sub(SubtypeInteger), []string{"precision"}, []string{"default_value"})
	register(FieldTypeNumber, sub(SubtypeDecimal), []string{"precision", "scale"}, []string{"default_value"})
	register(FieldTypeNumber, sub(SubtypeCurrency), nil, []string{"default_value"})
	register(FieldTypeNumber, sub(SubtypePercent), nil, []string{"default_value"})
	register(FieldTypeNumber, sub(SubtypeAutoNumber), []string{"format"}, []string{"start_value"})

	// boolean (no subtype)
	register(FieldTypeBoolean, nil, nil, []string{"default_value"})

	// datetime subtypes
	register(FieldTypeDatetime, sub(SubtypeDate), nil, []string{"default_value"})
	register(FieldTypeDatetime, sub(SubtypeDatetime), nil, []string{"default_value"})
	register(FieldTypeDatetime, sub(SubtypeTime), nil, []string{"default_value"})

	// picklist subtypes
	register(FieldTypePicklist, sub(SubtypeSingle), nil, []string{"values", "picklist_id", "default_value"})
	register(FieldTypePicklist, sub(SubtypeMulti), nil, []string{"values", "picklist_id", "default_value"})

	// reference subtypes
	register(FieldTypeReference, sub(SubtypeAssociation), []string{"relationship_name"}, []string{"on_delete"})
	register(FieldTypeReference, sub(SubtypeComposition), []string{"relationship_name"}, []string{"on_delete", "is_reparentable"})
	register(FieldTypeReference, sub(SubtypePolymorphic), []string{"relationship_name"}, nil)
}

func typeSubtypeKey(ft FieldType, fst *FieldSubtype) string {
	if fst == nil {
		return string(ft)
	}
	return fmt.Sprintf("%s/%s", ft, *fst)
}

// LookupTypeSpec returns the spec for a (type, subtype) pair, or false if invalid.
func LookupTypeSpec(ft FieldType, fst *FieldSubtype) (TypeSpec, bool) {
	key := typeSubtypeKey(ft, fst)
	spec, ok := typeRegistry[key]
	return spec, ok
}

// ValidTypeSubtypePairs returns all registered (type, subtype) pairs.
func ValidTypeSubtypePairs() []TypeSpec {
	result := make([]TypeSpec, 0, len(typeRegistry))
	for _, spec := range typeRegistry {
		result = append(result, spec)
	}
	return result
}
