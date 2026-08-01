package codegen

import (
	"strings"
	"unicode"
)

func tableVar(name string) string {
	if name == "" {
		return "Table"
	}
	return exportIdentifier(name)
}

func tableStruct(name string) string {
	return lowerFirst(tableVar(name)) + "Table"
}

func columnField(name string) string {
	parts := strings.Split(name, "_")
	for i, p := range parts {
		switch strings.ToLower(p) {
		case "id":
			parts[i] = "ID"
		case "url":
			parts[i] = "URL"
		case "uuid":
			parts[i] = "UUID"
		default:
			parts[i] = exportIdentifier(p)
		}
	}
	return strings.Join(parts, "")
}

func columnType(table, column string) string {
	return lowerFirst(tableVar(table)) + columnField(column) + "Col"
}

func exportIdentifier(s string) string {
	if s == "" {
		return s
	}
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		runes := []rune(p)
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, "")
}

func modelName(table string) string {
	return tableVar(table)
}

func enumConstName(typeName, label string) string {
	parts := strings.Split(label, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = exportIdentifier(p)
	}
	return typeName + strings.Join(parts, "")
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
