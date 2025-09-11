package utils

import "strings"

// FieldTypeAlias represents a mapping from user-friendly aliases to canonical types
type FieldTypeAlias struct {
	Alias         string // User input (e.g., "image", "belongsTo", "manyToMany")
	CanonicalType string // Standardized type (e.g., "storage.Attachment", "belongs_to", "many_to_many")
	GoType        string // Go type for struct fields
	Category      string // "storage", "relationship", "basic", "translation"
}

// FieldTypeAliases defines all supported field type aliases
var FieldTypeAliases = []FieldTypeAlias{
	// Storage types
	{"image", "storage.Attachment", "*storage.Attachment", "storage"},
	{"file", "storage.Attachment", "*storage.Attachment", "storage"},
	{"attachment", "storage.Attachment", "*storage.Attachment", "storage"},
	{"*storage.Attachment", "storage.Attachment", "*storage.Attachment", "storage"},

	// Translation types
	{"translation", "translation.Field", "translation.Field", "translation"},
	{"locale", "translation.Field", "translation.Field", "translation"},
	{"translatable", "translation.Field", "translation.Field", "translation"},
	{"translation.Field", "translation.Field", "translation.Field", "translation"},

	// Basic types with aliases
	{"text", "string", "string", "basic"},
	{"email", "string", "string", "basic"},
	{"password", "string", "string", "basic"},
	{"url", "string", "string", "basic"},
	{"phone", "string", "string", "basic"},

	// Relationship types - GORM standard names
	{"belongsTo", "belongs_to", "", "relationship"},
	{"belongs_to", "belongs_to", "", "relationship"},
	{"hasMany", "has_many", "", "relationship"},
	{"has_many", "has_many", "", "relationship"},
	{"hasOne", "has_one", "", "relationship"},
	{"has_one", "has_one", "", "relationship"},
	{"manyToMany", "many_to_many", "", "relationship"},
	{"many_to_many", "many_to_many", "", "relationship"},
	{"toMany", "many_to_many", "", "relationship"},
	{"to_many", "many_to_many", "", "relationship"},

	// Date/time aliases
	{"datetime", "types.DateTime", "types.DateTime", "basic"},
	{"date", "types.DateTime", "types.DateTime", "basic"},
	{"timestamp", "types.DateTime", "types.DateTime", "basic"},

	// JSON types
	{"json", "json.RawMessage", "json.RawMessage", "basic"},
	{"jsonb", "json.RawMessage", "json.RawMessage", "basic"},

	// Basic Go types
	{"string", "string", "string", "basic"},
	{"int", "int", "int", "basic"},
	{"uint", "uint", "uint", "basic"},
	{"bool", "bool", "bool", "basic"},
	{"boolean", "bool", "bool", "basic"},
	{"float", "float64", "float64", "basic"},
	{"float64", "float64", "float64", "basic"},
	{"time.Time", "time.Time", "time.Time", "basic"},
}

// ResolveFieldType resolves a field type alias to its canonical form
func ResolveFieldType(input string) FieldTypeAlias {
	// First, try exact match
	for _, alias := range FieldTypeAliases {
		if strings.EqualFold(alias.Alias, input) {
			return alias
		}
	}

	// If no match found, treat as custom type
	return FieldTypeAlias{
		Alias:         input,
		CanonicalType: input,
		GoType:        input,
		Category:      "custom",
	}
}

// IsRelationshipType checks if a type is a relationship
func IsRelationshipType(typeStr string) bool {
	resolved := ResolveFieldType(typeStr)
	return resolved.Category == "relationship"
}

// IsStorageType checks if a type is a storage attachment
func IsStorageType(typeStr string) bool {
	resolved := ResolveFieldType(typeStr)
	return resolved.Category == "storage"
}

// IsTranslationType checks if a type is a translation field
func IsTranslationType(typeStr string) bool {
	resolved := ResolveFieldType(typeStr)
	return resolved.Category == "translation"
}

// GetCanonicalRelationship returns the canonical relationship name
func GetCanonicalRelationship(input string) string {
	resolved := ResolveFieldType(input)
	if resolved.Category == "relationship" {
		return resolved.CanonicalType
	}
	return ""
}

// GetGoTypeFromAlias returns the Go type for a field type using the alias system
func GetGoTypeFromAlias(input string) string {
	resolved := ResolveFieldType(input)
	if resolved.GoType != "" {
		return resolved.GoType
	}
	return resolved.CanonicalType
}

// GetStorageType returns the storage type if applicable
func GetStorageType(input string) string {
	resolved := ResolveFieldType(input)
	if resolved.Category == "storage" {
		return resolved.CanonicalType
	}
	return ""
}

// GetTranslationType returns the translation type if applicable
func GetTranslationType(input string) string {
	resolved := ResolveFieldType(input)
	if resolved.Category == "translation" {
		return resolved.CanonicalType
	}
	return ""
}

// IsManyToManyRelationship checks if the type represents a many-to-many relationship
func IsManyToManyRelationship(typeStr string) bool {
	canonical := GetCanonicalRelationship(typeStr)
	return canonical == "many_to_many"
}

// IsBelongsToRelationship checks if the type represents a belongs_to relationship
func IsBelongsToRelationship(typeStr string) bool {
	canonical := GetCanonicalRelationship(typeStr)
	return canonical == "belongs_to"
}

// IsHasManyRelationship checks if the type represents a has_many relationship
func IsHasManyRelationship(typeStr string) bool {
	canonical := GetCanonicalRelationship(typeStr)
	return canonical == "has_many"
}

// IsHasOneRelationship checks if the type represents a has_one relationship
func IsHasOneRelationship(typeStr string) bool {
	canonical := GetCanonicalRelationship(typeStr)
	return canonical == "has_one"
}
