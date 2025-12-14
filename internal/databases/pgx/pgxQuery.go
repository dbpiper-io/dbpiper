package pgx

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

const AllTables = `
    SELECT table_name
    FROM information_schema.tables
    WHERE table_schema = 'public'
    ORDER BY table_name ASC
    `

const ColumnType = `
    SELECT column_name, data_type
    FROM information_schema.columns
    WHERE table_name = $1
  `


func SelectQuery(tableName string, fields []string) string {
	quotedTable := pgx.Identifier{tableName}.Sanitize()

	quotedFields := make([]string, len(fields))
	for i, field := range fields {
		quotedFields[i] = pgx.Identifier{field}.Sanitize()
	}

	return fmt.Sprintf("SELECT %s FROM %s",
		strings.Join(quotedFields, ", "),
		quotedTable)
}
