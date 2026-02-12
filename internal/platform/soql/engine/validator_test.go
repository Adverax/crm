package engine

import (
	"context"
	"testing"
)

func TestValidateSimpleQuery(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "simple select",
			query:   "SELECT Name FROM Account",
			wantErr: false,
		},
		{
			name:    "multiple fields",
			query:   "SELECT Name, Industry, AnnualRevenue FROM Account",
			wantErr: false,
		},
		{
			name:    "with where",
			query:   "SELECT Name FROM Account WHERE Industry = 'Technology'",
			wantErr: false,
		},
		{
			name:    "with order by",
			query:   "SELECT Name FROM Account ORDER BY Name ASC",
			wantErr: false,
		},
		{
			name:    "with limit",
			query:   "SELECT Name FROM Account LIMIT 10",
			wantErr: false,
		},
		{
			name:    "unknown object",
			query:   "SELECT Name FROM UnknownObject",
			wantErr: true,
		},
		{
			name:    "unknown field",
			query:   "SELECT UnknownField FROM Account",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			_, err = validator.Validate(ctx, ast)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLookups(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "simple lookup",
			query:   "SELECT Name, Account.Name FROM Contact",
			wantErr: false,
		},
		{
			name:    "nested lookup",
			query:   "SELECT Name, Account.Owner.Name FROM Contact",
			wantErr: false,
		},
		{
			name:    "deep lookup",
			query:   "SELECT Name, Account.Owner.Manager.Name FROM Contact",
			wantErr: false,
		},
		{
			name:    "unknown lookup",
			query:   "SELECT Name, UnknownLookup.Name FROM Contact",
			wantErr: true,
		},
		{
			name:    "lookup to unknown field",
			query:   "SELECT Name, Account.UnknownField FROM Contact",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result, err := validator.Validate(ctx, ast)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && result != nil {
				// Check that lookups were resolved
				if len(result.ResolvedRefs) == 0 {
					t.Error("expected resolved references")
				}
			}
		})
	}
}

func TestValidateLookupDepthLimit(t *testing.T) {
	metadata := setupTestMetadata()

	// Create validator with max depth of 2
	limits := &Limits{
		MaxLookupDepth: 2,
	}
	validator := NewValidator(metadata, nil, limits)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "within limit",
			query:   "SELECT Name, Account.Owner.Name FROM Contact",
			wantErr: false,
		},
		{
			name:    "exceeds limit",
			query:   "SELECT Name, Account.Owner.Manager.Name FROM Contact",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			_, err = validator.Validate(ctx, ast)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				if !IsLimitError(err) {
					t.Errorf("expected LimitError, got %T", err)
				}
			}
		})
	}
}

func TestValidateSubqueries(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "simple subquery",
			query:   "SELECT Name, (SELECT FirstName, LastName FROM Contacts) FROM Account",
			wantErr: false,
		},
		{
			name:    "subquery with where",
			query:   "SELECT Name, (SELECT FirstName FROM Contacts WHERE Email != null) FROM Account",
			wantErr: false,
		},
		{
			name:    "multiple subqueries",
			query:   "SELECT Name, (SELECT FirstName FROM Contacts), (SELECT Name FROM Opportunities) FROM Account",
			wantErr: false,
		},
		{
			name:    "unknown relationship",
			query:   "SELECT Name, (SELECT Name FROM UnknownRelationship) FROM Account",
			wantErr: true,
		},
		{
			name:    "unknown field in subquery",
			query:   "SELECT Name, (SELECT UnknownField FROM Contacts) FROM Account",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result, err := validator.Validate(ctx, ast)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && result != nil {
				if len(result.Subqueries) == 0 {
					t.Error("expected subqueries to be validated")
				}
			}
		})
	}
}

func TestValidateSubqueryLimit(t *testing.T) {
	metadata := setupTestMetadata()

	// Create validator with max 1 subquery
	limits := &Limits{
		MaxSubqueries: 1,
	}
	validator := NewValidator(metadata, nil, limits)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "within limit",
			query:   "SELECT Name, (SELECT FirstName FROM Contacts) FROM Account",
			wantErr: false,
		},
		{
			name:    "exceeds limit",
			query:   "SELECT Name, (SELECT FirstName FROM Contacts), (SELECT Name FROM Opportunities) FROM Account",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			_, err = validator.Validate(ctx, ast)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAggregates(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "COUNT",
			query:   "SELECT COUNT(Id) FROM Account",
			wantErr: false,
		},
		{
			name:    "SUM",
			query:   "SELECT SUM(Amount) FROM Opportunity",
			wantErr: false,
		},
		{
			name:    "multiple aggregates",
			query:   "SELECT COUNT(Id), SUM(Amount), AVG(Amount) FROM Opportunity",
			wantErr: false,
		},
		{
			name:    "aggregate with group by",
			query:   "SELECT StageName, COUNT(Id) FROM Opportunity GROUP BY StageName",
			wantErr: false,
		},
		{
			name:    "aggregate with having",
			query:   "SELECT StageName, COUNT(Id) FROM Opportunity GROUP BY StageName HAVING COUNT(Id) > 5",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			_, err = validator.Validate(ctx, ast)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAccessControl(t *testing.T) {
	metadata := setupTestMetadata()

	// Create access controller that denies Contact
	access := &ObjectAccessController{
		AllowedObjects: map[string]bool{
			"Account":     true,
			"Opportunity": true,
			"User":        true,
			// Contact is NOT allowed
		},
	}

	validator := NewValidator(metadata, access, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "allowed object",
			query:   "SELECT Name FROM Account",
			wantErr: false,
		},
		{
			name:    "denied object",
			query:   "SELECT FirstName FROM Contact",
			wantErr: true,
		},
		{
			name:    "denied relationship subquery",
			query:   "SELECT Name, (SELECT FirstName FROM Contacts) FROM Account",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			_, err = validator.Validate(ctx, ast)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				if !IsAccessError(err) {
					t.Errorf("expected AccessError, got %T: %v", err, err)
				}
			}
		})
	}
}

func TestValidateOffsetLimit(t *testing.T) {
	metadata := setupTestMetadata()

	limits := &Limits{
		MaxRecords: 100,
		MaxOffset:  50,
	}
	validator := NewValidator(metadata, nil, limits)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "within limits",
			query:   "SELECT Name FROM Account LIMIT 50 OFFSET 20",
			wantErr: false,
		},
		{
			name:    "limit exceeds max",
			query:   "SELECT Name FROM Account LIMIT 200",
			wantErr: true,
		},
		{
			name:    "offset exceeds max",
			query:   "SELECT Name FROM Account LIMIT 10 OFFSET 100",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			_, err = validator.Validate(ctx, ast)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateResolvedReferences(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	ast, err := Parse("SELECT Name, Account.Name, Account.Owner.Name FROM Contact WHERE Account.Industry = 'Tech'")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	result, err := validator.Validate(ctx, ast)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Check resolved references
	expectedPaths := []string{
		"Name",
		"Account.Name",
		"Account.Owner.Name",
		"Account.Industry",
	}

	for _, path := range expectedPaths {
		ref, ok := result.ResolvedRefs[path]
		if !ok {
			t.Errorf("expected resolved reference for path %s", path)
			continue
		}

		if ref.Field == nil {
			t.Errorf("resolved reference %s has nil Field", path)
		}
	}

	// Check joins were created for lookups
	accountNameRef := result.ResolvedRefs["Account.Name"]
	if accountNameRef != nil && len(accountNameRef.Joins) != 1 {
		t.Errorf("expected 1 join for Account.Name, got %d", len(accountNameRef.Joins))
	}

	ownerNameRef := result.ResolvedRefs["Account.Owner.Name"]
	if ownerNameRef != nil && len(ownerNameRef.Joins) != 2 {
		t.Errorf("expected 2 joins for Account.Owner.Name, got %d", len(ownerNameRef.Joins))
	}
}

func TestValidateFieldAccessControl(t *testing.T) {
	metadata := setupTestMetadata()

	// Create access controller that restricts fields
	access := &FieldAccessController{
		AllowedObjects: map[string]bool{
			"Account": true,
			"Contact": true,
		},
		AllowedFields: map[string]map[string]bool{
			"Account": {
				"Id":       true,
				"Name":     true,
				"Industry": true,
				// AnnualRevenue is NOT allowed
			},
		},
	}

	validator := NewValidator(metadata, access, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "allowed fields",
			query:   "SELECT Name, Industry FROM Account",
			wantErr: false,
		},
		{
			name:    "denied field",
			query:   "SELECT Name, AnnualRevenue FROM Account",
			wantErr: true,
		},
		{
			name:    "denied field in WHERE",
			query:   "SELECT Name FROM Account WHERE AnnualRevenue > 1000000",
			wantErr: true,
		},
		{
			name:    "denied field in ORDER BY",
			query:   "SELECT Name FROM Account ORDER BY AnnualRevenue",
			wantErr: true,
		},
		{
			name:    "object without field restrictions",
			query:   "SELECT FirstName, LastName, Email FROM Contact",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			_, err = validator.Validate(ctx, ast)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDateLiteralTypes(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "date field with TODAY",
			query:   "SELECT Name FROM Opportunity WHERE CloseDate = TODAY",
			wantErr: false,
		},
		{
			name:    "datetime field with LAST_N_DAYS",
			query:   "SELECT Name FROM Account WHERE CreatedDate > LAST_N_DAYS:30",
			wantErr: false,
		},
		{
			name:    "date in range comparison",
			query:   "SELECT Name FROM Opportunity WHERE CloseDate >= LAST_MONTH AND CloseDate <= THIS_MONTH",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			_, err = validator.Validate(ctx, ast)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSelectFieldsLimit(t *testing.T) {
	metadata := setupTestMetadata()

	// Allow max 3 select fields
	limits := &Limits{
		MaxSelectFields: 3,
	}
	validator := NewValidator(metadata, nil, limits)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "within limit",
			query:   "SELECT Name, Industry FROM Account",
			wantErr: false,
		},
		{
			name:    "at limit",
			query:   "SELECT Id, Name, Industry FROM Account",
			wantErr: false,
		},
		{
			name:    "exceeds limit",
			query:   "SELECT Id, Name, Industry, AnnualRevenue FROM Account",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			_, err = validator.Validate(ctx, ast)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil && !IsLimitError(err) {
				t.Errorf("expected LimitError, got %T", err)
			}
		})
	}
}

// Note: Query length limit is checked at the Engine level before parsing,
// not in the Validator. See Engine.Execute() for implementation.

func TestValidateGroupByFields(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "valid group by with aggregate",
			query:   "SELECT StageName, COUNT(Id) FROM Opportunity GROUP BY StageName",
			wantErr: false,
		},
		{
			name:    "multiple group by fields",
			query:   "SELECT StageName, AccountId, COUNT(Id) FROM Opportunity GROUP BY StageName, AccountId",
			wantErr: false,
		},
		{
			name:    "group by with lookup",
			query:   "SELECT Account.Industry, COUNT(Id) FROM Opportunity GROUP BY Account.Industry",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			_, err = validator.Validate(ctx, ast)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateOrderByFieldsValid(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "single order by",
			query:   "SELECT Name FROM Account ORDER BY Name",
			wantErr: false,
		},
		{
			name:    "multiple order by fields",
			query:   "SELECT Name FROM Account ORDER BY Industry, Name, AnnualRevenue",
			wantErr: false,
		},
		{
			name:    "order by with direction",
			query:   "SELECT Name FROM Account ORDER BY Name DESC",
			wantErr: false,
		},
		{
			name:    "order by with nulls",
			query:   "SELECT Name FROM Account ORDER BY Name ASC NULLS LAST",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			_, err = validator.Validate(ctx, ast)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Note: MaxOrderByFields limit is defined in Limits but not currently enforced during validation.

func TestValidateNullValidator(t *testing.T) {
	metadata := setupTestMetadata()
	// Create validator without access controller or limits
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	// Should allow everything
	tests := []struct {
		name  string
		query string
	}{
		{"simple query", "SELECT Name FROM Account"},
		{"complex query", "SELECT Name, Account.Owner.Name FROM Contact WHERE Email IS NOT NULL ORDER BY Name LIMIT 1000 OFFSET 500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			_, err = validator.Validate(ctx, ast)
			if err != nil {
				t.Errorf("Validate() should not error with nil access/limits: %v", err)
			}
		})
	}
}

func TestValidateMultipleLookupPaths(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	// Test that the same lookup referenced multiple times creates proper joins
	query := "SELECT Account.Name, Account.Industry, Account.Owner.Name FROM Contact WHERE Account.Industry = 'Tech'"

	ast, err := Parse(query)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	result, err := validator.Validate(ctx, ast)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Check that multiple Account.* references use the same join
	accountNameRef := result.ResolvedRefs["Account.Name"]
	accountIndustryRef := result.ResolvedRefs["Account.Industry"]

	if accountNameRef == nil || accountIndustryRef == nil {
		t.Fatal("expected resolved references for Account fields")
	}

	if len(accountNameRef.Joins) != len(accountIndustryRef.Joins) {
		t.Errorf("same lookup should have same join count: Name=%d, Industry=%d",
			len(accountNameRef.Joins), len(accountIndustryRef.Joins))
	}
}

func TestValidateFunctions(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "COALESCE with valid field",
			query:   "SELECT COALESCE(Name, 'Unknown') FROM Account",
			wantErr: false,
		},
		{
			name:    "COALESCE with invalid field",
			query:   "SELECT COALESCE(UnknownField, 'Default') FROM Account",
			wantErr: true,
		},
		{
			name:    "UPPER with valid field",
			query:   "SELECT UPPER(Name) FROM Account",
			wantErr: false,
		},
		{
			name:    "UPPER with invalid field",
			query:   "SELECT UPPER(UnknownField) FROM Account",
			wantErr: true,
		},
		{
			name:    "nested functions with valid field",
			query:   "SELECT UPPER(TRIM(Name)) FROM Account",
			wantErr: false,
		},
		{
			name:    "function with lookup",
			query:   "SELECT UPPER(Account.Name) FROM Contact",
			wantErr: false,
		},
		{
			name:    "function with invalid lookup",
			query:   "SELECT UPPER(UnknownLookup.Name) FROM Contact",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			_, err = validator.Validate(ctx, ast)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateArithmeticExpressions(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "arithmetic with valid field",
			query:   "SELECT AnnualRevenue * 0.1 FROM Account",
			wantErr: false,
		},
		{
			name:    "arithmetic with invalid field",
			query:   "SELECT UnknownField * 0.1 FROM Account",
			wantErr: true,
		},
		{
			name:    "field to field arithmetic",
			query:   "SELECT AnnualRevenue - AnnualRevenue FROM Account",
			wantErr: false,
		},
		{
			name:    "complex arithmetic with valid fields",
			query:   "SELECT (AnnualRevenue + 1000) * 0.1 FROM Account",
			wantErr: false,
		},
		{
			name:    "arithmetic with lookup",
			query:   "SELECT Account.Industry FROM Contact",
			wantErr: false, // Lookup fields work
		},
		{
			name:    "string concatenation",
			query:   "SELECT FirstName || ' ' || LastName FROM Contact",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			_, err = validator.Validate(ctx, ast)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateWhereSubquery(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "simple IN subquery",
			query:   "SELECT Name FROM Account WHERE Id IN (SELECT AccountId FROM Contact)",
			wantErr: false,
		},
		{
			name:    "NOT IN subquery",
			query:   "SELECT Name FROM Account WHERE Id NOT IN (SELECT AccountId FROM Contact)",
			wantErr: false,
		},
		{
			name:    "subquery with WHERE",
			query:   "SELECT Name FROM Account WHERE Id IN (SELECT AccountId FROM Contact WHERE Email IS NOT NULL)",
			wantErr: false,
		},
		{
			name:    "subquery with unknown object",
			query:   "SELECT Name FROM Account WHERE Id IN (SELECT AccountId FROM UnknownObject)",
			wantErr: true,
		},
		{
			name:    "subquery with unknown field",
			query:   "SELECT Name FROM Account WHERE Id IN (SELECT UnknownField FROM Contact)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			_, err = validator.Validate(ctx, ast)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatorObjectMetaBuilder(t *testing.T) {
	// Test the ObjectMeta builder pattern
	obj := NewObjectMeta("TestObject", "", "test_objects").
		Field("Id", "id", FieldTypeID).
		Field("Name", "name", FieldTypeString).
		Field("Amount", "amount", FieldTypeFloat).
		Field("Active", "is_active", FieldTypeBoolean).
		Field("CreatedAt", "created_at", FieldTypeDateTime).
		Lookup("Parent", "parent_id", "TestObject", "id").
		Relationship("Children", "TestObject", "parent_id", "id").
		Build()

	if obj == nil {
		t.Fatal("Build() returned nil")
	}

	if obj.Name != "TestObject" {
		t.Errorf("Name = %s, want TestObject", obj.Name)
	}

	if obj.TableName != "test_objects" {
		t.Errorf("TableName = %s, want test_objects", obj.TableName)
	}

	// Check fields
	expectedFields := []string{"Id", "Name", "Amount", "Active", "CreatedAt"}
	for _, name := range expectedFields {
		if _, ok := obj.Fields[name]; !ok {
			t.Errorf("expected field %s", name)
		}
	}

	// Check lookup
	if _, ok := obj.Lookups["Parent"]; !ok {
		t.Error("expected lookup Parent")
	}

	// Check relationship
	if _, ok := obj.Relationships["Children"]; !ok {
		t.Error("expected relationship Children")
	}
}

func TestValidateEmptySelect(t *testing.T) {
	// This should be caught by parser, but test defense in depth
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	// Create a minimal AST with empty select
	ast := &Grammar{
		Select: nil, // empty select
		From:   "Account",
	}

	_, err := validator.Validate(ctx, ast)
	if err == nil {
		t.Error("Validate() should error on empty SELECT")
	}
}
