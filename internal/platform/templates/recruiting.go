package templates

import "github.com/adverax/crm/internal/platform/metadata"

// Recruiting returns the Recruiting application template.
func Recruiting() Template {
	assoc := metadata.SubtypeAssociation
	plain := metadata.SubtypePlain
	area := metadata.SubtypeArea
	email := metadata.SubtypeEmail
	phone := metadata.SubtypePhone
	urlSub := metadata.SubtypeURL
	integer := metadata.SubtypeInteger
	currency := metadata.SubtypeCurrency
	date := metadata.SubtypeDate
	datetime := metadata.SubtypeDatetime
	single := metadata.SubtypeSingle

	onDeleteSetNull := "set_null"
	onDeleteCascade := "cascade"

	maxLen255 := 255
	maxLen100 := 100
	maxLen50 := 50
	maxLen2000 := 2000

	return Template{
		ID:          "recruiting",
		Label:       "Recruiting",
		Description: "Applicant tracking system: positions, candidates, applications, and interviews",
		Objects: []ObjectTemplate{
			{
				APIName:               "position",
				Label:                 "Position",
				PluralLabel:           "Positions",
				Description:           "Open job positions",
				Visibility:            metadata.VisibilityPublicRead,
				IsCreateable:          true,
				IsUpdateable:          true,
				IsDeleteable:          true,
				IsQueryable:           true,
				IsSearchable:          true,
				IsCustomFieldsAllowed: true,
				HasNotes:              true,
				HasSharingRules:       true,
			},
			{
				APIName:               "candidate",
				Label:                 "Candidate",
				PluralLabel:           "Candidates",
				Description:           "Job applicants",
				Visibility:            metadata.VisibilityPrivate,
				IsCreateable:          true,
				IsUpdateable:          true,
				IsDeleteable:          true,
				IsQueryable:           true,
				IsSearchable:          true,
				IsCustomFieldsAllowed: true,
				HasNotes:              true,
				HasSharingRules:       true,
			},
			{
				APIName:               "application",
				Label:                 "Application",
				PluralLabel:           "Applications",
				Description:           "Candidate applications for positions",
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
				APIName:               "interview",
				Label:                 "Interview",
				PluralLabel:           "Interviews",
				Description:           "Scheduled interviews for applications",
				Visibility:            metadata.VisibilityPrivate,
				IsCreateable:          true,
				IsUpdateable:          true,
				IsDeleteable:          true,
				IsQueryable:           true,
				IsSearchable:          true,
				IsCustomFieldsAllowed: true,
				HasNotes:              true,
				HasSharingRules:       true,
			},
		},
		Fields: []FieldTemplate{
			// --- Position fields ---
			{ObjectAPIName: "position", APIName: "title", Label: "Title", FieldType: metadata.FieldTypeText, FieldSubtype: &plain, IsRequired: true, Config: metadata.FieldConfig{MaxLength: &maxLen255}, SortOrder: 1},
			{ObjectAPIName: "position", APIName: "department", Label: "Department", FieldType: metadata.FieldTypeText, FieldSubtype: &plain, Config: metadata.FieldConfig{MaxLength: &maxLen100}, SortOrder: 2},
			{ObjectAPIName: "position", APIName: "location", Label: "Location", FieldType: metadata.FieldTypeText, FieldSubtype: &plain, Config: metadata.FieldConfig{MaxLength: &maxLen255}, SortOrder: 3},
			{ObjectAPIName: "position", APIName: "status", Label: "Status", FieldType: metadata.FieldTypePicklist, FieldSubtype: &single, IsRequired: true, Config: metadata.FieldConfig{MaxLength: &maxLen50}, SortOrder: 4},
			{ObjectAPIName: "position", APIName: "salary_min", Label: "Salary Min", FieldType: metadata.FieldTypeNumber, FieldSubtype: &currency, Config: metadata.FieldConfig{Precision: intPtr(18), Scale: intPtr(2)}, SortOrder: 5},
			{ObjectAPIName: "position", APIName: "salary_max", Label: "Salary Max", FieldType: metadata.FieldTypeNumber, FieldSubtype: &currency, Config: metadata.FieldConfig{Precision: intPtr(18), Scale: intPtr(2)}, SortOrder: 6},
			{ObjectAPIName: "position", APIName: "headcount", Label: "Headcount", FieldType: metadata.FieldTypeNumber, FieldSubtype: &integer, SortOrder: 7},
			{ObjectAPIName: "position", APIName: "description", Label: "Description", FieldType: metadata.FieldTypeText, FieldSubtype: &area, Config: metadata.FieldConfig{MaxLength: &maxLen2000}, SortOrder: 8},

			// --- Candidate fields ---
			{ObjectAPIName: "candidate", APIName: "first_name", Label: "First Name", FieldType: metadata.FieldTypeText, FieldSubtype: &plain, IsRequired: true, Config: metadata.FieldConfig{MaxLength: &maxLen100}, SortOrder: 1},
			{ObjectAPIName: "candidate", APIName: "last_name", Label: "Last Name", FieldType: metadata.FieldTypeText, FieldSubtype: &plain, IsRequired: true, Config: metadata.FieldConfig{MaxLength: &maxLen100}, SortOrder: 2},
			{ObjectAPIName: "candidate", APIName: "email", Label: "Email", FieldType: metadata.FieldTypeText, FieldSubtype: &email, Config: metadata.FieldConfig{MaxLength: &maxLen255}, SortOrder: 3},
			{ObjectAPIName: "candidate", APIName: "phone", Label: "Phone", FieldType: metadata.FieldTypeText, FieldSubtype: &phone, Config: metadata.FieldConfig{MaxLength: &maxLen50}, SortOrder: 4},
			{ObjectAPIName: "candidate", APIName: "current_company", Label: "Current Company", FieldType: metadata.FieldTypeText, FieldSubtype: &plain, Config: metadata.FieldConfig{MaxLength: &maxLen255}, SortOrder: 5},
			{ObjectAPIName: "candidate", APIName: "current_title", Label: "Current Title", FieldType: metadata.FieldTypeText, FieldSubtype: &plain, Config: metadata.FieldConfig{MaxLength: &maxLen255}, SortOrder: 6},
			{ObjectAPIName: "candidate", APIName: "linkedin_url", Label: "LinkedIn URL", FieldType: metadata.FieldTypeText, FieldSubtype: &urlSub, Config: metadata.FieldConfig{MaxLength: &maxLen255}, SortOrder: 7},
			{ObjectAPIName: "candidate", APIName: "source", Label: "Source", FieldType: metadata.FieldTypePicklist, FieldSubtype: &single, Config: metadata.FieldConfig{MaxLength: &maxLen100}, SortOrder: 8},
			{ObjectAPIName: "candidate", APIName: "notes", Label: "Notes", FieldType: metadata.FieldTypeText, FieldSubtype: &area, Config: metadata.FieldConfig{MaxLength: &maxLen2000}, SortOrder: 9},

			// --- Application fields ---
			{ObjectAPIName: "application", APIName: "position_id", Label: "Position", FieldType: metadata.FieldTypeReference, FieldSubtype: &assoc, ReferencedObjectAPIName: "position", IsRequired: true, Config: metadata.FieldConfig{OnDelete: &onDeleteSetNull, RelationshipName: strPtr("position")}, SortOrder: 1},
			{ObjectAPIName: "application", APIName: "candidate_id", Label: "Candidate", FieldType: metadata.FieldTypeReference, FieldSubtype: &assoc, ReferencedObjectAPIName: "candidate", IsRequired: true, Config: metadata.FieldConfig{OnDelete: &onDeleteSetNull, RelationshipName: strPtr("candidate")}, SortOrder: 2},
			{ObjectAPIName: "application", APIName: "stage", Label: "Stage", FieldType: metadata.FieldTypePicklist, FieldSubtype: &single, IsRequired: true, Config: metadata.FieldConfig{MaxLength: &maxLen50}, SortOrder: 3},
			{ObjectAPIName: "application", APIName: "applied_date", Label: "Applied Date", FieldType: metadata.FieldTypeDatetime, FieldSubtype: &date, SortOrder: 4},
			{ObjectAPIName: "application", APIName: "is_rejected", Label: "Is Rejected", FieldType: metadata.FieldTypeBoolean, SortOrder: 5},
			{ObjectAPIName: "application", APIName: "rejection_reason", Label: "Rejection Reason", FieldType: metadata.FieldTypeText, FieldSubtype: &area, Config: metadata.FieldConfig{MaxLength: &maxLen2000}, SortOrder: 6},

			// --- Interview fields ---
			{ObjectAPIName: "interview", APIName: "application_id", Label: "Application", FieldType: metadata.FieldTypeReference, FieldSubtype: &assoc, ReferencedObjectAPIName: "application", IsRequired: true, Config: metadata.FieldConfig{OnDelete: &onDeleteCascade, RelationshipName: strPtr("application")}, SortOrder: 1},
			{ObjectAPIName: "interview", APIName: "scheduled_at", Label: "Scheduled At", FieldType: metadata.FieldTypeDatetime, FieldSubtype: &datetime, IsRequired: true, SortOrder: 2},
			{ObjectAPIName: "interview", APIName: "interview_type", Label: "Interview Type", FieldType: metadata.FieldTypePicklist, FieldSubtype: &single, Config: metadata.FieldConfig{MaxLength: &maxLen50}, SortOrder: 3},
			{ObjectAPIName: "interview", APIName: "result", Label: "Result", FieldType: metadata.FieldTypePicklist, FieldSubtype: &single, Config: metadata.FieldConfig{MaxLength: &maxLen50}, SortOrder: 4},
			{ObjectAPIName: "interview", APIName: "feedback", Label: "Feedback", FieldType: metadata.FieldTypeText, FieldSubtype: &area, Config: metadata.FieldConfig{MaxLength: &maxLen2000}, SortOrder: 5},
		},
	}
}
