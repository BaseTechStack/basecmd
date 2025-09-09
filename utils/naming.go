package utils

import (
	"strings"
	"unicode"
)

// NamingConvention holds all naming variations derived from a single model name
type NamingConvention struct {
	// Original input (e.g., "ProductCategory")
	Original string

	// Model naming
	Model      string // ProductCategory (PascalCase)
	ModelLower string // productCategory (camelCase)
	ModelSnake string // product_category (snake_case)
	ModelKebab string // product-category (kebab-case)

	// Plural forms
	Plural      string // ProductCategories (PascalCase)
	PluralLower string // productCategories (camelCase)
	PluralSnake string // product_categories (snake_case)
	PluralKebab string // product-categories (kebab-case)

	// Package and directory naming
	PackageName string // product_categories (snake_case plural for package)
	DirName     string // product_categories (snake_case plural for directory)

	// Route naming
	RoutePath  string // /product-categories (kebab-case plural)
	RouteGroup string // product-categories (kebab-case plural)

	// Controller and Service naming
	Controller string // ProductCategoriesController
	Service    string // ProductCategoriesService

	// Database naming
	TableName string // product_categories (snake_case plural)

	// Variable naming
	VarSingle string // productCategory (camelCase)
	VarPlural string // productCategories (camelCase)
	VarId     string // productCategoryId (camelCase + Id)
}

// NewNamingConvention creates all naming variations from a single model name
func NewNamingConvention(modelName string) *NamingConvention {
	// Ensure PascalCase for the model name
	model := ToPascalCase(modelName)
	plural := PluralizeClient.Plural(model)

	nc := &NamingConvention{
		Original: modelName,

		// Model naming
		Model:      model,
		ModelLower: ToCamelCase(model),
		ModelSnake: ToSnakeCase(model),
		ModelKebab: ToKebabCase(model),

		// Plural forms
		Plural:      plural,
		PluralLower: ToCamelCase(plural),
		PluralSnake: ToSnakeCase(plural),
		PluralKebab: ToKebabCase(plural),

		// Package and directory naming
		PackageName: ToSnakeCase(plural),
		DirName:     ToSnakeCase(plural),

		// Route naming (kebab-case is URL-friendly)
		RoutePath:  "/" + ToKebabCase(plural),
		RouteGroup: ToKebabCase(plural),

		// Controller and Service naming (singular like model)
		Controller: model + "Controller",
		Service:    model + "Service",

		// Database naming
		TableName: ToSnakeCase(plural),

		// Variable naming
		VarSingle: ToCamelCase(model),
		VarPlural: ToCamelCase(plural),
		VarId:     ToCamelCase(model) + "Id",
	}

	return nc
}

// Field represents a clean field structure - keeping compatibility with templates
type Field struct {
	Name    string // Field name in Go (PascalCase)
	Type    string // Go type
	JSONTag string // JSON tag name (maps to JSONName for template compatibility)
	GORMTag string // Complete GORM tag (maps to GORM for template compatibility)

	// Template compatibility fields
	JSONName           string // Same as JSONTag for template compatibility
	DBName             string // Database column name
	GORM               string // Same as GORMTag for template compatibility
	Relationship       string // Same as RelationType for template compatibility
	RelatedModel       string // Related model name (PascalCase)
	ForeignKey         string // Foreign key field name
	TestValue          string // Test value for this field
	UpdateTestValue    string // Update test value (maps to UpdateValue)
	TestValueWithIndex string // Test value with index for loops
	TestValueUnique    string // Unique test value for constraint tests

	// For relations
	IsRelation   bool
	RelationType string // belongs_to, has_many, has_one, many_to_many

	// Validation
	IsRequired bool
	IsUnique   bool

	// Special types
	IsImage      bool
	IsFile       bool
	IsAttachment bool
	IsMedia      bool
}

// ParseField creates a properly structured Field from a field definition string
func ParseField(fieldDef string) Field {
	parts := strings.Split(fieldDef, ":")
	fieldName := parts[0]
	var fieldType string

	// Smart field inference: if only field name provided, infer type
	if len(parts) == 1 {
		fieldType = inferFieldType(fieldName)
	} else {
		fieldType = parts[1]
	}

	field := Field{
		Name:    ToPascalCase(fieldName),
		JSONTag: ToSnakeCase(fieldName),
	}

	// Set template compatibility fields
	field.JSONName = field.JSONTag
	field.DBName = field.JSONTag
	field.Relationship = ""
	field.IsRelation = false

	// Handle relationships using alias system
	if IsRelationshipType(fieldType) {
		canonical := GetCanonicalRelationship(fieldType)
		switch canonical {
		case "belongs_to":
			return parseBelongsToField(fieldName, parts, field)
		case "has_many":
			return parseHasManyField(fieldName, parts, field)
		case "has_one":
			return parseHasOneField(fieldName, parts, field)
		case "many_to_many":
			return parseManyToManyField(fieldName, parts, field)
		}
	} else if fieldType == "attachment" || fieldType == "file" || fieldType == "image" {
		return parseAttachmentField(fieldName, fieldType, field)
	}

	// Handle regular fields using the new alias system
	resolved := ResolveFieldType(fieldType)
	field.Type = GetGoTypeFromAlias(fieldType)

	// Handle special categories
	switch resolved.Category {
	case "storage":
		field.JSONTag = ToSnakeCase(fieldName) + ",omitempty"
		field.JSONName = ToSnakeCase(fieldName) + ",omitempty"
		field.GORMTag = `gorm:"foreignKey:ModelId;references:Id"`
	case "translation":
		field.Type = resolved.GoType
		field.JSONTag = ToSnakeCase(fieldName)
		field.JSONName = ToSnakeCase(fieldName)
		field.GORMTag = `gorm:"foreignKey:ModelId;references:Id"`
	}

	field.GORM = field.GORMTag

	return field
}

// parseBelongsToField handles belongsTo relationship fields
func parseBelongsToField(fieldName string, parts []string, field Field) Field {
	field.IsRelation = true
	field.RelationType = "belongs_to"
	field.Relationship = "belongs_to"

	var relatedModel string
	if len(parts) > 2 {
		relatedModel = strings.TrimSpace(parts[2])
	} else {
		// Auto-detect from field name (remove _id suffix if present)
		baseName := strings.TrimSuffix(fieldName, "_id")
		relatedModel = ToPascalCase(baseName)
	}

	// For belongsTo, we need the foreign key field (ends with Id)
	// If field name already ends with _id or Id, use it as is, otherwise add _id
	var foreignKeyName string
	lowerFieldName := strings.ToLower(fieldName)
	if strings.HasSuffix(lowerFieldName, "_id") || strings.HasSuffix(lowerFieldName, "id") {
		foreignKeyName = fieldName
	} else {
		foreignKeyName = fieldName + "_id"
	}

	field.Name = ToPascalCase(foreignKeyName)
	field.Type = "uint"
	field.JSONTag = ToSnakeCase(foreignKeyName)
	field.JSONName = field.JSONTag
	field.DBName = field.JSONTag
	field.RelatedModel = relatedModel
	field.ForeignKey = field.Name

	return field
}

// parseHasManyField handles hasMany relationship fields
func parseHasManyField(fieldName string, parts []string, field Field) Field {
	field.IsRelation = true
	field.RelationType = "has_many"
	field.Relationship = "has_many"

	var relatedModel string
	if len(parts) > 2 {
		relatedModel = ToPascalCase(parts[2])
	} else {
		// Infer model from field name (plural to singular)
		relatedModel = ToPascalCase(Singularize(fieldName))
	}

	field.Type = "[]*" + relatedModel
	field.RelatedModel = relatedModel
	field.GORM = field.GORMTag

	return field
}

// parseHasOneField handles hasOne relationship fields
func parseHasOneField(fieldName string, parts []string, field Field) Field {
	field.IsRelation = true
	field.RelationType = "has_one"
	field.Relationship = "has_one"

	var relatedModel string
	if len(parts) > 2 {
		relatedModel = ToPascalCase(parts[2])
	} else {
		relatedModel = ToPascalCase(fieldName)
	}

	field.Type = "*" + relatedModel
	field.RelatedModel = relatedModel
	field.GORM = field.GORMTag

	return field
}

// parseManyToManyField handles manyToMany relationship fields
func parseManyToManyField(fieldName string, parts []string, field Field) Field {
	field.IsRelation = true
	field.RelationType = "many_to_many"
	field.Relationship = "many_to_many"

	var relatedModel string
	if len(parts) > 2 {
		relatedModel = ToPascalCase(parts[2])
	} else {
		// Infer model from field name (plural to singular)
		relatedModel = ToPascalCase(Singularize(fieldName))
	}

	field.Type = "[]*" + relatedModel
	field.RelatedModel = relatedModel
	field.GORM = field.GORMTag

	return field
}

// parseAttachmentField handles attachment/file/image fields
func parseAttachmentField(_ string, fieldType string, field Field) Field {
	field.Type = "*storage.Attachment"
	field.IsAttachment = true

	switch fieldType {
	case "file":
		field.IsFile = true
	case "image":
		field.IsImage = true
	}

	field.GORM = `gorm:"foreignKey:ModelId;references:Id"`
	field.GORMTag = field.GORM

	return field
}

// inferFieldType infers the Go type from field name patterns
func inferFieldType(fieldName string) string {
	fieldName = strings.ToLower(fieldName)

	// Check for common patterns
	if strings.HasSuffix(fieldName, "_id") {
		return "uint"
	}
	if strings.HasSuffix(fieldName, "_at") || strings.HasSuffix(fieldName, "_date") || strings.HasSuffix(fieldName, "_time") {
		return "time.Time"
	}
	if strings.HasSuffix(fieldName, "_count") || strings.HasSuffix(fieldName, "_number") {
		return "int"
	}
	if strings.Contains(fieldName, "email") {
		return "string"
	}
	if strings.Contains(fieldName, "phone") {
		return "string"
	}
	if strings.Contains(fieldName, "url") || strings.Contains(fieldName, "link") {
		return "string"
	}
	if strings.Contains(fieldName, "description") || strings.Contains(fieldName, "content") || strings.Contains(fieldName, "text") {
		return "text"
	}
	if strings.Contains(fieldName, "active") || strings.Contains(fieldName, "enabled") || strings.Contains(fieldName, "published") {
		return "bool"
	}

	// Default to string
	return "string"
}

// ToCapitalCase converts snake_case or kebab-case to Capital Case
func ToCapitalCase(s string) string {
	// Replace underscores and hyphens with spaces
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	// Split by spaces and capitalize each word
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}

// SplitCamelCase splits a CamelCase or PascalCase string into words
func SplitCamelCase(s string) []string {
	var words []string
	var currentWord []rune

	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			// Check if next character is lowercase (indicates new word)
			if i+1 < len(s) {
				nextRune := rune(s[i+1])
				if unicode.IsLower(nextRune) || (len(currentWord) > 1 && unicode.IsUpper(currentWord[len(currentWord)-1])) {
					if len(currentWord) > 0 {
						words = append(words, string(currentWord))
						currentWord = []rune{}
					}
				}
			} else if len(currentWord) > 0 {
				// Last character
				words = append(words, string(currentWord))
				currentWord = []rune{}
			}
		}
		currentWord = append(currentWord, r)
	}

	if len(currentWord) > 0 {
		words = append(words, string(currentWord))
	}

	return words
}
