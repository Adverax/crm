package templates

import "github.com/adverax/crm/internal/platform/metadata"

// SalesCRM returns the Sales CRM application template.
func SalesCRM() Template {
	assoc := metadata.SubtypeAssociation
	plain := metadata.SubtypePlain
	area := metadata.SubtypeArea
	email := metadata.SubtypeEmail
	phone := metadata.SubtypePhone
	urlSub := metadata.SubtypeURL
	integer := metadata.SubtypeInteger
	currency := metadata.SubtypeCurrency
	percent := metadata.SubtypePercent
	date := metadata.SubtypeDate
	single := metadata.SubtypeSingle

	onDeleteSetNull := "set_null"

	maxLen255 := 255
	maxLen100 := 100
	maxLen50 := 50
	maxLen2000 := 2000

	return Template{
		ID:          "sales_crm",
		Label:       "Sales CRM",
		Description: "CRM for sales teams: accounts, contacts, opportunities, and tasks",
		Objects: []ObjectTemplate{
			{
				APIName:               "account",
				Label:                 "Account",
				PluralLabel:           "Accounts",
				Description:           "Companies and organizations",
				Visibility:            metadata.VisibilityPrivate,
				IsCreateable:          true,
				IsUpdateable:          true,
				IsDeleteable:          true,
				IsQueryable:           true,
				IsSearchable:          true,
				IsCustomFieldsAllowed: true,
				HasActivities:         true,
				HasNotes:              true,
				HasSharingRules:       true,
			},
			{
				APIName:               "contact",
				Label:                 "Contact",
				PluralLabel:           "Contacts",
				Description:           "People associated with accounts",
				Visibility:            metadata.VisibilityPrivate,
				IsCreateable:          true,
				IsUpdateable:          true,
				IsDeleteable:          true,
				IsQueryable:           true,
				IsSearchable:          true,
				IsCustomFieldsAllowed: true,
				HasActivities:         true,
				HasNotes:              true,
				HasSharingRules:       true,
			},
			{
				APIName:               "opportunity",
				Label:                 "Opportunity",
				PluralLabel:           "Opportunities",
				Description:           "Sales deals and pipeline",
				Visibility:            metadata.VisibilityPrivate,
				IsCreateable:          true,
				IsUpdateable:          true,
				IsDeleteable:          true,
				IsQueryable:           true,
				IsSearchable:          true,
				IsCustomFieldsAllowed: true,
				HasActivities:         true,
				HasNotes:              true,
				HasSharingRules:       true,
			},
			{
				APIName:               "task",
				Label:                 "Task",
				PluralLabel:           "Tasks",
				Description:           "Activities and to-do items",
				Visibility:            metadata.VisibilityPublicReadWrite,
				IsCreateable:          true,
				IsUpdateable:          true,
				IsDeleteable:          true,
				IsQueryable:           true,
				IsSearchable:          true,
				IsCustomFieldsAllowed: true,
				HasActivities:         false,
				HasNotes:              true,
				HasSharingRules:       true,
			},
		},
		Fields: []FieldTemplate{
			// --- Account fields ---
			{ObjectAPIName: "account", APIName: "name", Label: "Name", FieldType: metadata.FieldTypeText, FieldSubtype: &plain, IsRequired: true, Config: metadata.FieldConfig{MaxLength: &maxLen255}, SortOrder: 1},
			{ObjectAPIName: "account", APIName: "website", Label: "Website", FieldType: metadata.FieldTypeText, FieldSubtype: &urlSub, Config: metadata.FieldConfig{MaxLength: &maxLen255}, SortOrder: 2},
			{ObjectAPIName: "account", APIName: "phone", Label: "Phone", FieldType: metadata.FieldTypeText, FieldSubtype: &phone, Config: metadata.FieldConfig{MaxLength: &maxLen50}, SortOrder: 3},
			{ObjectAPIName: "account", APIName: "industry", Label: "Industry", FieldType: metadata.FieldTypePicklist, FieldSubtype: &single, Config: metadata.FieldConfig{MaxLength: &maxLen100}, SortOrder: 4},
			{ObjectAPIName: "account", APIName: "employee_count", Label: "Employee Count", FieldType: metadata.FieldTypeNumber, FieldSubtype: &integer, SortOrder: 5},
			{ObjectAPIName: "account", APIName: "annual_revenue", Label: "Annual Revenue", FieldType: metadata.FieldTypeNumber, FieldSubtype: &currency, Config: metadata.FieldConfig{Precision: intPtr(18), Scale: intPtr(2)}, SortOrder: 6},
			{ObjectAPIName: "account", APIName: "billing_city", Label: "Billing City", FieldType: metadata.FieldTypeText, FieldSubtype: &plain, Config: metadata.FieldConfig{MaxLength: &maxLen100}, SortOrder: 7},
			{ObjectAPIName: "account", APIName: "billing_country", Label: "Billing Country", FieldType: metadata.FieldTypeText, FieldSubtype: &plain, Config: metadata.FieldConfig{MaxLength: &maxLen100}, SortOrder: 8},
			{ObjectAPIName: "account", APIName: "description", Label: "Description", FieldType: metadata.FieldTypeText, FieldSubtype: &area, Config: metadata.FieldConfig{MaxLength: &maxLen2000}, SortOrder: 9},

			// --- Contact fields ---
			{ObjectAPIName: "contact", APIName: "first_name", Label: "First Name", FieldType: metadata.FieldTypeText, FieldSubtype: &plain, IsRequired: true, Config: metadata.FieldConfig{MaxLength: &maxLen100}, SortOrder: 1},
			{ObjectAPIName: "contact", APIName: "last_name", Label: "Last Name", FieldType: metadata.FieldTypeText, FieldSubtype: &plain, IsRequired: true, Config: metadata.FieldConfig{MaxLength: &maxLen100}, SortOrder: 2},
			{ObjectAPIName: "contact", APIName: "email", Label: "Email", FieldType: metadata.FieldTypeText, FieldSubtype: &email, Config: metadata.FieldConfig{MaxLength: &maxLen255}, SortOrder: 3},
			{ObjectAPIName: "contact", APIName: "phone", Label: "Phone", FieldType: metadata.FieldTypeText, FieldSubtype: &phone, Config: metadata.FieldConfig{MaxLength: &maxLen50}, SortOrder: 4},
			{ObjectAPIName: "contact", APIName: "title", Label: "Title", FieldType: metadata.FieldTypeText, FieldSubtype: &plain, Config: metadata.FieldConfig{MaxLength: &maxLen255}, SortOrder: 5},
			{ObjectAPIName: "contact", APIName: "account_id", Label: "Account", FieldType: metadata.FieldTypeReference, FieldSubtype: &assoc, ReferencedObjectAPIName: "account", Config: metadata.FieldConfig{OnDelete: &onDeleteSetNull, RelationshipName: strPtr("account")}, SortOrder: 6},
			{ObjectAPIName: "contact", APIName: "department", Label: "Department", FieldType: metadata.FieldTypeText, FieldSubtype: &plain, Config: metadata.FieldConfig{MaxLength: &maxLen100}, SortOrder: 7},
			{ObjectAPIName: "contact", APIName: "date_of_birth", Label: "Date of Birth", FieldType: metadata.FieldTypeDatetime, FieldSubtype: &date, SortOrder: 8},
			{ObjectAPIName: "contact", APIName: "description", Label: "Description", FieldType: metadata.FieldTypeText, FieldSubtype: &area, Config: metadata.FieldConfig{MaxLength: &maxLen2000}, SortOrder: 9},

			// --- Opportunity fields ---
			{ObjectAPIName: "opportunity", APIName: "name", Label: "Name", FieldType: metadata.FieldTypeText, FieldSubtype: &plain, IsRequired: true, Config: metadata.FieldConfig{MaxLength: &maxLen255}, SortOrder: 1},
			{ObjectAPIName: "opportunity", APIName: "account_id", Label: "Account", FieldType: metadata.FieldTypeReference, FieldSubtype: &assoc, ReferencedObjectAPIName: "account", Config: metadata.FieldConfig{OnDelete: &onDeleteSetNull, RelationshipName: strPtr("account")}, SortOrder: 2},
			{ObjectAPIName: "opportunity", APIName: "contact_id", Label: "Contact", FieldType: metadata.FieldTypeReference, FieldSubtype: &assoc, ReferencedObjectAPIName: "contact", Config: metadata.FieldConfig{OnDelete: &onDeleteSetNull, RelationshipName: strPtr("contact")}, SortOrder: 3},
			{ObjectAPIName: "opportunity", APIName: "amount", Label: "Amount", FieldType: metadata.FieldTypeNumber, FieldSubtype: &currency, Config: metadata.FieldConfig{Precision: intPtr(18), Scale: intPtr(2)}, SortOrder: 4},
			{ObjectAPIName: "opportunity", APIName: "stage", Label: "Stage", FieldType: metadata.FieldTypePicklist, FieldSubtype: &single, IsRequired: true, Config: metadata.FieldConfig{MaxLength: &maxLen100}, SortOrder: 5},
			{ObjectAPIName: "opportunity", APIName: "probability", Label: "Probability", FieldType: metadata.FieldTypeNumber, FieldSubtype: &percent, Config: metadata.FieldConfig{Precision: intPtr(5), Scale: intPtr(2)}, SortOrder: 6},
			{ObjectAPIName: "opportunity", APIName: "close_date", Label: "Close Date", FieldType: metadata.FieldTypeDatetime, FieldSubtype: &date, SortOrder: 7},
			{ObjectAPIName: "opportunity", APIName: "is_won", Label: "Is Won", FieldType: metadata.FieldTypeBoolean, SortOrder: 8},
			{ObjectAPIName: "opportunity", APIName: "description", Label: "Description", FieldType: metadata.FieldTypeText, FieldSubtype: &area, Config: metadata.FieldConfig{MaxLength: &maxLen2000}, SortOrder: 9},

			// --- Task fields ---
			{ObjectAPIName: "task", APIName: "subject", Label: "Subject", FieldType: metadata.FieldTypeText, FieldSubtype: &plain, IsRequired: true, Config: metadata.FieldConfig{MaxLength: &maxLen255}, SortOrder: 1},
			{ObjectAPIName: "task", APIName: "status", Label: "Status", FieldType: metadata.FieldTypePicklist, FieldSubtype: &single, IsRequired: true, Config: metadata.FieldConfig{MaxLength: &maxLen50}, SortOrder: 2},
			{ObjectAPIName: "task", APIName: "priority", Label: "Priority", FieldType: metadata.FieldTypePicklist, FieldSubtype: &single, Config: metadata.FieldConfig{MaxLength: &maxLen50}, SortOrder: 3},
			{ObjectAPIName: "task", APIName: "due_date", Label: "Due Date", FieldType: metadata.FieldTypeDatetime, FieldSubtype: &date, SortOrder: 4},
			{ObjectAPIName: "task", APIName: "account_id", Label: "Account", FieldType: metadata.FieldTypeReference, FieldSubtype: &assoc, ReferencedObjectAPIName: "account", Config: metadata.FieldConfig{OnDelete: &onDeleteSetNull, RelationshipName: strPtr("account")}, SortOrder: 5},
			{ObjectAPIName: "task", APIName: "contact_id", Label: "Contact", FieldType: metadata.FieldTypeReference, FieldSubtype: &assoc, ReferencedObjectAPIName: "contact", Config: metadata.FieldConfig{OnDelete: &onDeleteSetNull, RelationshipName: strPtr("contact")}, SortOrder: 6},
			{ObjectAPIName: "task", APIName: "opportunity_id", Label: "Opportunity", FieldType: metadata.FieldTypeReference, FieldSubtype: &assoc, ReferencedObjectAPIName: "opportunity", Config: metadata.FieldConfig{OnDelete: &onDeleteSetNull, RelationshipName: strPtr("opportunity")}, SortOrder: 7},
			{ObjectAPIName: "task", APIName: "is_completed", Label: "Is Completed", FieldType: metadata.FieldTypeBoolean, SortOrder: 8},
			{ObjectAPIName: "task", APIName: "description", Label: "Description", FieldType: metadata.FieldTypeText, FieldSubtype: &area, Config: metadata.FieldConfig{MaxLength: &maxLen2000}, SortOrder: 9},
		},
	}
}

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }
