package utils

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/model.tmpl
var modelTemplate string

//go:embed templates/controller.tmpl
var controllerTemplate string

//go:embed templates/service.tmpl
var serviceTemplate string

//go:embed templates/module.tmpl
var moduleTemplate string

//go:embed templates/validator.tmpl
var validatorTemplate string

// TemplateData contains all data needed for template generation
type TemplateData struct {
	// Naming conventions for the model
	*NamingConvention

	// Fields including relations
	Fields []Field

	// Computed properties
	HasRelations          bool
	HasBelongsTo          bool
	HasHasMany            bool
	HasHasOne             bool
	HasManyToMany         bool
	HasImages             bool
	HasFiles              bool
	HasMedia              bool
	HasAttachments        bool
	HasTimestamps         bool
	HasSoftDelete         bool
	HasTranslatableFields bool

	// Import paths needed
	Imports []string

	// Join tables for many-to-many relationships
	JoinTables []string
}

// NewTemplateData creates template data from model name and field definitions
func NewTemplateData(modelName string, fieldDefs []string) *TemplateData {
	nc := NewNamingConvention(modelName)
	td := &TemplateData{
		NamingConvention: nc,
		Fields:           []Field{},
		Imports:          []string{},
	}

	// Generate field structs using centralized parsing
	for _, fieldDef := range fieldDefs {
		field := ParseField(fieldDef)

		// Handle belongsTo relationships - need both foreign key and relationship object
		if field.Relationship == "belongs_to" {
			// Add the foreign key field
			td.Fields = append(td.Fields, field)

			// Add the relationship object field
			objectName := TrimIdSuffix(field.Name)
			relationField := Field{
				Name:         objectName,
				Type:         "*" + field.RelatedModel,
				JSONTag:      ToSnakeCase(objectName) + ",omitempty",
				JSONName:     ToSnakeCase(objectName) + ",omitempty",
				DBName:       ToSnakeCase(objectName),
				GORM:         fmt.Sprintf(`gorm:"foreignKey:%s"`, field.Name),
				GORMTag:      fmt.Sprintf(`gorm:"foreignKey:%s"`, field.Name),
				Relationship: "belongs_to_object",
				RelatedModel: field.RelatedModel,
				IsRelation:   true,
				RelationType: "belongs_to_object",
			}
			td.Fields = append(td.Fields, relationField)
		} else {
			td.Fields = append(td.Fields, field)
		}

		// Update computed properties
		td.updateComputedProperties(field)
	}

	// Add standard imports
	td.addStandardImports()

	return td
}

// updateComputedProperties updates computed properties based on field
func (td *TemplateData) updateComputedProperties(field Field) {
	if field.IsRelation {
		td.HasRelations = true
		switch field.RelationType {
		case "belongs_to":
			td.HasBelongsTo = true
		case "has_many":
			td.HasHasMany = true
		case "has_one":
			td.HasHasOne = true
		case "many_to_many":
			td.HasManyToMany = true
		}
	}

	if field.IsImage {
		td.HasImages = true
		td.HasAttachments = true
	}
	if field.IsFile {
		td.HasFiles = true
		td.HasAttachments = true
	}
	if field.IsMedia {
		td.HasMedia = true
		td.HasAttachments = true
	}
	// Check for translatable fields
	if field.Type == "translation.Field" {
		td.HasTranslatableFields = true
	}
	if field.Type == "media.Media" {
		td.HasMedia = true
	}
	if field.Type == "time.Time" {
		switch field.Name {
		case "DeletedAt":
			td.HasSoftDelete = true
		case "CreatedAt", "UpdatedAt":
			td.HasTimestamps = true
		}
	}
}

// addStandardImports adds standard imports based on fields
func (td *TemplateData) addStandardImports() {
	imports := make(map[string]bool)

	// Always needed
	imports["time"] = true
	imports["gorm.io/gorm"] = true

	// Check fields for additional imports
	for _, field := range td.Fields {
		switch field.Type {
		case "time.Time":
			imports["time"] = true
		case "datatypes.JSON":
			imports["gorm.io/datatypes"] = true
		case "*storage.Attachment":
			imports["base/core/storage"] = true
		case "translation.Field":
			imports["base/core/translation"] = true
		case "media.Media":
			imports["base/core/media"] = true
		}
	}

	// Convert map to slice
	td.Imports = []string{}
	for imp := range imports {
		td.Imports = append(td.Imports, imp)
	}
}

// HasFieldType checks if any field has the specified type
func HasFieldType(fields []Field, fieldType string) bool {
	for _, field := range fields {
		if field.Type == fieldType {
			return true
		}
	}
	return false
}

// HasImageField checks if any field has image type
func HasImageField(fields []Field) bool {
	return HasFieldType(fields, "*storage.Attachment")
}

// HasMediaField checks if any field has media type
func HasMediaField(fields []Field) bool {
	for _, field := range fields {
		if field.Type == "media.Media" {
			return true
		}
	}
	return false
}

// Singularize converts plural to singular (basic implementation)
func Singularize(word string) string {
	if strings.HasSuffix(word, "ies") {
		return strings.TrimSuffix(word, "ies") + "y"
	}
	if strings.HasSuffix(word, "es") {
		return strings.TrimSuffix(word, "es")
	}
	if strings.HasSuffix(word, "s") {
		return strings.TrimSuffix(word, "s")
	}
	return word
}

// GenerateFieldStructs processes all fields and returns a slice of Field (for backward compatibility)
func GenerateFieldStructs(fieldDefs []string) []Field {
	var fields []Field
	for _, fieldDef := range fieldDefs {
		field := parseFieldDef(fieldDef)
		fields = append(fields, field)
	}
	return fields
}

// parseFieldDef parses a single field definition using the new alias system
func parseFieldDef(fieldDef string) Field {
	parts := strings.Split(fieldDef, ":")
	fieldName := parts[0]
	var fieldType string

	if len(parts) == 1 {
		fieldType = inferFieldTypeCompat(fieldName)
	} else if len(parts) == 2 {
		fieldType = parts[1]
	} else {
		// For relationship definitions like "tags:toMany:Tag", pass the full definition
		fieldType = fieldDef
	}

	field := Field{
		Name:     ToPascalCase(fieldName),
		JSONName: ToSnakeCase(fieldName),
		DBName:   ToSnakeCase(fieldName),
	}

	// Handle relationship definitions with multiple parts
	if strings.Contains(fieldType, ":") {
		parts := strings.Split(fieldType, ":")
		if len(parts) >= 2 {
			relationType := parts[1]

			// Use alias system to resolve relationship type
			if IsRelationshipType(relationType) {
				canonical := GetCanonicalRelationship(relationType)
				field.IsRelation = true
				field.Relationship = canonical
				field.RelationType = canonical

				// Set related model
				if len(parts) > 2 {
					field.RelatedModel = ToPascalCase(parts[2])
				} else {
					field.RelatedModel = ToPascalCase(Singularize(fieldName))
				}

				// Set appropriate type and tags based on relationship
				switch canonical {
				case "many_to_many":
					field.Type = "[]*" + field.RelatedModel
					field.JSONName = ToSnakeCase(fieldName)
				case "has_many":
					field.Type = "[]*" + field.RelatedModel
					field.JSONName = ToSnakeCase(fieldName) + ",omitempty"
				case "has_one":
					field.Type = "*" + field.RelatedModel
					field.JSONName = ToSnakeCase(fieldName) + ",omitempty"
				case "belongs_to":
					field.Type = "*" + field.RelatedModel
					field.JSONName = ToSnakeCase(fieldName) + ",omitempty"
				}
			}
		}
	} else {
		// Single field type - use alias system
		resolved := ResolveFieldType(fieldType)
		field.Type = GetGoTypeFromAlias(fieldType)

		// Handle special categories
		switch resolved.Category {
		case "storage":
			field.JSONName = ToSnakeCase(fieldName) + ",omitempty"
			field.GORMTag = `gorm:"foreignKey:ModelId;references:Id"`
		case "media":
			field.JSONName = ToSnakeCase(fieldName) + ",omitempty"
			field.GORMTag = `gorm:"foreignKey:ModelId;references:Id"`
		case "translation":
			// Translation fields are stored as translation.Field and handled like storage attachments
			field.Type = resolved.GoType
			field.JSONName = ToSnakeCase(fieldName)
			field.GORMTag = `gorm:"foreignKey:ModelId;references:Id"`
		}
	}

	// Set compatibility fields
	field.JSONTag = field.JSONName
	field.GORM = field.GORMTag

	return field
}

// Helper functions for backward compatibility
func inferFieldTypeCompat(fieldName string) string {
	lower := strings.ToLower(fieldName)
	if strings.HasSuffix(lower, "_id") {
		return "uint"
	}
	return "string"
}

// GenerateFileFromTemplate generates a file from embedded template (for backward compatibility)
func GenerateFileFromTemplate(dir, filename, templateName string, naming *NamingConvention, fields []Field) {
	// Convert Field slice to embedded template data
	var tmplContent string
	switch templateName {
	case "model.tmpl":
		tmplContent = modelTemplate
	case "controller.tmpl":
		tmplContent = controllerTemplate
	case "service.tmpl":
		tmplContent = serviceTemplate
	case "module.tmpl":
		tmplContent = moduleTemplate
	case "validator.tmpl":
		tmplContent = validatorTemplate
	default:
		fmt.Printf("Unknown template: %s\n", templateName)
		return
	}

	// Create template with functions
	funcMap := template.FuncMap{
		"toLower":      strings.ToLower,
		"toTitle":      ToTitle,
		"ToSnakeCase":  ToSnakeCase,
		"ToPascalCase": ToPascalCase,
		"ToKebabCase":  ToKebabCase,
		"ToPlural":     ToPlural,
		"TrimIdSuffix": TrimIdSuffix,
		"hasPrefix":    strings.HasPrefix,
		"hasSuffix":    strings.HasSuffix,
		"contains":     strings.Contains,
		"eq":           func(a, b interface{}) bool { return a == b },
		"slice": func(s string, start, end int) string {
			if start >= len(s) {
				return ""
			}
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
		"hasField": func(fields []Field, fieldType string) bool {
			return HasFieldType(fields, fieldType)
		},
	}

	tmpl, err := template.New(templateName).Funcs(funcMap).Parse(tmplContent)
	if err != nil {
		fmt.Printf("Error parsing template %s: %v\n", templateName, err)
		return
	}

	// Create output directory
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	// Create output file
	outputFile := filepath.Join(dir, filename)
	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", outputFile, err)
		return
	}
	defer f.Close()

	// Execute template with data structure
	data := struct {
		*NamingConvention
		Fields                []Field
		HasImageField         bool
		HasMediaField         bool
		HasTranslatableFields bool
		HasSoftDelete         bool
		HasTimestamps         bool
		HasAttachments        bool
		HasRelations          bool
		HasBelongsTo          bool
		HasHasMany            bool
		HasHasOne             bool
		HasManyToMany         bool
	}{
		NamingConvention:      naming,
		Fields:                fields,
		HasImageField:         HasImageField(fields),
		HasMediaField:         HasMediaField(fields),
		HasTranslatableFields: HasFieldType(fields, "translation.Field"),
		HasSoftDelete:         HasFieldType(fields, "gorm.DeletedAt"),
		HasTimestamps:         HasFieldType(fields, "time.Time"),
		HasAttachments:        HasFieldType(fields, "*storage.Attachment"),
		HasRelations:          HasFieldType(fields, "*models."),
		HasBelongsTo:          HasFieldType(fields, "belongsTo"),
		HasHasMany:            HasFieldType(fields, "hasMany"),
		HasHasOne:             HasFieldType(fields, "hasOne"),
		HasManyToMany:         HasFieldType(fields, "manyToMany"),
	}

	if err := tmpl.Execute(f, data); err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		return
	}

	fmt.Printf("Generated %s\n", outputFile)
}
