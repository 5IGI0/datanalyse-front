package main

import (
	"encoding/json"
	"fmt"
)

type GeneratorData struct {
	Format       string   `json:"format"`
	PrimaryType  string   `json:"primary_type"`
	Tags         []string `json:"tags"`
	Version      int      `json:"ver"`
	LinkedColumn string   `json:"linked_col"`
	CustomData   any      `json:"custom_data"`
}

type GeneratorInfo struct {
	Name          string `json:"name"`
	VersionString string `json:"ver"`
	VersionId     uint32 `json:"ver_id"`
}

type AnalyzerManifest struct {
	Analyzer GeneratorInfo `json:"analyzer"`
	Data     any           `json:"data"`
}

type ColumnInfo struct {
	Generator     *GeneratorInfo `json:"generator"`
	GeneratorData *GeneratorData `json:"generator_data"`
	IsInvisible   bool           `json:"is_invisible"`
	Tags          []string       `json:"tags"`
	Analyzers     map[string]AnalyzerManifest
}

type TableDescription struct {
	tableName     string
	Name          string              `json:"name"`
	Description   string              `json:"description"`
	Version       uint32              `json:"version"`
	Meta          map[string]string   `json:"meta"`
	GroupAnalyzer []AnalyzerManifest  `json:"group_analyzers"`
	VirtColMap    map[string][]string `json:"virtcol_map"`

	columns map[string]ColumnInfo
	/*	since MySQL doesn't support reversed string indexes
		we need another column to reverse them */
	reverse_columns map[string]string

	TableInfo struct {
		RowCount int `json:"row_count"`
		Size     int `json:"size"`
	} `json:"table_info"`
}

func DescribeTable(table_name string) (TableDescription, error) {
	var table TableDescription
	var table_info struct {
		RowCount int    `db:"TABLE_ROWS"`
		Size     int    `db:"DATA_LENGTH"`
		Comment  string `db:"TABLE_COMMENT"`
	}

	/* get table info */
	AssertError(
		GlobalContext.Database.QueryRowx(`
			SELECT TABLE_ROWS, DATA_LENGTH, TABLE_COMMENT
			FROM INFORMATION_SCHEMA.TABLES
			WHERE TABLE_SCHEMA=? AND TABLE_NAME=?`, GlobalContext.DatabaseName, table_name).
			StructScan(&table_info))

	if table_info.Comment == "" {
		return table, fmt.Errorf("no comment for table `%s`", table_name)
	}

	AssertError(json.Unmarshal([]byte(table_info.Comment), &table))

	table.tableName = table_name
	table.TableInfo.Size = table_info.Size
	table.TableInfo.RowCount = table_info.RowCount

	/* get column infos */
	table.columns = make(map[string]ColumnInfo)
	table.reverse_columns = make(map[string]string)

	rows, err := GlobalContext.Database.Queryx(`
			SELECT COLUMN_NAME, COLUMN_COMMENT
			FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_SCHEMA=? AND TABLE_NAME=?`, GlobalContext.DatabaseName, table_name)
	AssertError(err)
	defer rows.Close()

	for rows.Next() {
		var colinf ColumnInfo
		var Column struct {
			Name    string `db:"COLUMN_NAME"`
			Comment string `db:"COLUMN_COMMENT"`
		}

		AssertError(rows.StructScan(&Column))

		if Column.Comment != "" {
			AssertError(json.Unmarshal([]byte(Column.Comment), &colinf))
			table.columns[Column.Name] = colinf

			if colinf.Generator != nil && colinf.Generator.Name == "reverse_index_emulator" {
				table.reverse_columns[colinf.GeneratorData.LinkedColumn] = Column.Name
			}
		}
	}

	return table, nil
}
