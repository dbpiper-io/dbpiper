package pgx

const AllTables = `
    SELECT table_name
    FROM information_schema.tables
    WHERE table_schema = 'public'
    ORDER BY table_name ASC
    `

func GetColumnType(table string) (string, []any) {
	sql := `
        SELECT column_name, data_type
        FROM information_schema.columns
        WHERE table_name = $1
    `
	return sql, []any{table}
}
