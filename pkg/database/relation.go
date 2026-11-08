package database

import (
	"fmt"
	"strings"
)

// Singularize provides basic singularization for table names (e.g., "users" -> "user", "articles" -> "article").
func Singularize(name string) string {
	lower := strings.ToLower(name)
	if strings.HasSuffix(lower, "ies") {
		return strings.TrimSuffix(lower, "ies") + "y"
	}
	if strings.HasSuffix(lower, "es") && (strings.HasSuffix(lower, "shes") || strings.HasSuffix(lower, "ches") || strings.HasSuffix(lower, "sses")) {
		return strings.TrimSuffix(lower, "es")
	}
	if strings.HasSuffix(lower, "s") && !strings.HasSuffix(lower, "ss") {
		return strings.TrimSuffix(lower, "s")
	}
	return lower
}

// Pluralize provides basic pluralization for table names (e.g., "user" -> "users", "category" -> "categories").
func Pluralize(name string) string {
	lower := strings.ToLower(name)
	if strings.HasSuffix(lower, "y") && len(lower) > 1 && !strings.ContainsAny(string(lower[len(lower)-2]), "aeiou") {
		return strings.TrimSuffix(lower, "y") + "ies"
	}
	if strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "x") || strings.HasSuffix(lower, "ch") || strings.HasSuffix(lower, "sh") {
		return lower + "es"
	}
	return lower + "s"
}

// HasMany builds a query for records in relTable that belong to parentRow.
// If foreignKey is empty, defaults to Singularize(parentTable) + "_id" (e.g. "user_id").
// If localKey is empty, defaults to "id".
func HasMany(conn *Connection, parentTable string, parentRow map[string]any, relTable, foreignKey, localKey string) (*QueryBuilder, error) {
	if localKey == "" {
		localKey = "id"
	}
	if parentTable == "" {
		if t, ok := parentRow["_tabel"].(string); ok && t != "" {
			parentTable = t
		} else if t, ok := parentRow["tabel"].(string); ok && t != "" {
			parentTable = t
		} else if t, ok := parentRow["table"].(string); ok && t != "" {
			parentTable = t
		}
	}
	if foreignKey == "" {
		if parentTable != "" {
			foreignKey = Singularize(parentTable) + "_id"
		} else {
			foreignKey = "user_id"
		}
	}
	localVal, ok := parentRow[localKey]
	if !ok {
		return nil, fmt.Errorf("local key %q not found in parent row", localKey)
	}

	qb := conn.Table(relTable).Where(foreignKey, "=", localVal)
	return qb, nil
}

// BelongsTo builds a query for the parent record in parentTable that childRow belongs to.
// If foreignKey is empty, defaults to Singularize(parentTable) + "_id" (e.g. "user_id").
// If ownerKey is empty, defaults to "id".
func BelongsTo(conn *Connection, childRow map[string]any, parentTable, foreignKey, ownerKey string) (*QueryBuilder, error) {
	if ownerKey == "" {
		ownerKey = "id"
	}
	if foreignKey == "" {
		foreignKey = Singularize(parentTable) + "_id"
	}
	foreignVal, ok := childRow[foreignKey]
	if !ok {
		return nil, fmt.Errorf("foreign key %q not found in child row", foreignKey)
	}

	qb := conn.Table(parentTable).Where(ownerKey, "=", foreignVal)
	return qb, nil
}

// HasOne builds a query for a single related record in relTable that belongs to parentRow.
func HasOne(conn *Connection, parentTable string, parentRow map[string]any, relTable, foreignKey, localKey string) (*QueryBuilder, error) {
	qb, err := HasMany(conn, parentTable, parentRow, relTable, foreignKey, localKey)
	if err != nil {
		return nil, err
	}
	qb.Limit(1)
	return qb, nil
}
