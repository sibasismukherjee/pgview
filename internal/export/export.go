// Package export writes query results to CSV or JSON files.
package export

import (
	"encoding/csv"
	"encoding/json"
	"io"
)

// nullSentinel is the string value produced by db.formatValue when a cell is NULL.
const nullSentinel = "NULL"

// WriteCSV writes cols as a header row followed by rows to w.
// NULL values are written as empty strings.
func WriteCSV(w io.Writer, cols []string, rows [][]string) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(cols); err != nil {
		return err
	}
	for _, row := range rows {
		out := make([]string, len(row))
		for i, v := range row {
			if v == nullSentinel {
				out[i] = ""
			} else {
				out[i] = v
			}
		}
		if err := cw.Write(out); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// WriteJSON writes an indented JSON array of objects to w.
// NULL values are written as JSON null.
func WriteJSON(w io.Writer, cols []string, rows [][]string) error {
	records := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		obj := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			if i < len(row) {
				if row[i] == nullSentinel {
					obj[col] = nil
				} else {
					obj[col] = row[i]
				}
			}
		}
		records = append(records, obj)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(records)
}
