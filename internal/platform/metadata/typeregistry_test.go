package metadata

import (
	"testing"
)

func TestLookupTypeSpec(t *testing.T) {
	t.Parallel()

	sub := func(s FieldSubtype) *FieldSubtype { return &s }

	tests := []struct {
		name   string
		ft     FieldType
		fst    *FieldSubtype
		wantOK bool
	}{
		{name: "text/plain is valid", ft: FieldTypeText, fst: sub(SubtypePlain), wantOK: true},
		{name: "text/email is valid", ft: FieldTypeText, fst: sub(SubtypeEmail), wantOK: true},
		{name: "text/area is valid", ft: FieldTypeText, fst: sub(SubtypeArea), wantOK: true},
		{name: "text/rich is valid", ft: FieldTypeText, fst: sub(SubtypeRich), wantOK: true},
		{name: "text/phone is valid", ft: FieldTypeText, fst: sub(SubtypePhone), wantOK: true},
		{name: "text/url is valid", ft: FieldTypeText, fst: sub(SubtypeURL), wantOK: true},
		{name: "number/integer is valid", ft: FieldTypeNumber, fst: sub(SubtypeInteger), wantOK: true},
		{name: "number/decimal is valid", ft: FieldTypeNumber, fst: sub(SubtypeDecimal), wantOK: true},
		{name: "number/currency is valid", ft: FieldTypeNumber, fst: sub(SubtypeCurrency), wantOK: true},
		{name: "number/percent is valid", ft: FieldTypeNumber, fst: sub(SubtypePercent), wantOK: true},
		{name: "number/auto_number is valid", ft: FieldTypeNumber, fst: sub(SubtypeAutoNumber), wantOK: true},
		{name: "boolean/null is valid", ft: FieldTypeBoolean, fst: nil, wantOK: true},
		{name: "datetime/date is valid", ft: FieldTypeDatetime, fst: sub(SubtypeDate), wantOK: true},
		{name: "datetime/datetime is valid", ft: FieldTypeDatetime, fst: sub(SubtypeDatetime), wantOK: true},
		{name: "datetime/time is valid", ft: FieldTypeDatetime, fst: sub(SubtypeTime), wantOK: true},
		{name: "picklist/single is valid", ft: FieldTypePicklist, fst: sub(SubtypeSingle), wantOK: true},
		{name: "picklist/multi is valid", ft: FieldTypePicklist, fst: sub(SubtypeMulti), wantOK: true},
		{name: "reference/association is valid", ft: FieldTypeReference, fst: sub(SubtypeAssociation), wantOK: true},
		{name: "reference/composition is valid", ft: FieldTypeReference, fst: sub(SubtypeComposition), wantOK: true},
		{name: "reference/polymorphic is valid", ft: FieldTypeReference, fst: sub(SubtypePolymorphic), wantOK: true},
		{name: "text/null is invalid", ft: FieldTypeText, fst: nil, wantOK: false},
		{name: "boolean/plain is invalid", ft: FieldTypeBoolean, fst: sub(SubtypePlain), wantOK: false},
		{name: "text/integer is invalid", ft: FieldTypeText, fst: sub(SubtypeInteger), wantOK: false},
		{name: "unknown/anything is invalid", ft: "unknown", fst: sub(SubtypePlain), wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, ok := LookupTypeSpec(tt.ft, tt.fst)
			if ok != tt.wantOK {
				t.Errorf("LookupTypeSpec(%s, %v) ok = %v, want %v", tt.ft, tt.fst, ok, tt.wantOK)
			}
		})
	}
}

func TestValidTypeSubtypePairs(t *testing.T) {
	t.Parallel()
	pairs := ValidTypeSubtypePairs()
	if len(pairs) != 20 {
		t.Errorf("ValidTypeSubtypePairs() returned %d pairs, want 20", len(pairs))
	}
}
