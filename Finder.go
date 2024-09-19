package main

import (
	"slices"

	"github.com/Masterminds/squirrel"
)

type SearchableColumn struct {
	Dataset  string   `json:"dataset"`
	Columns  []string `json:"column"`
	ColKey   string   `json:"colkey"`
	Features []string `json:"feature"`
	// 0 - unlikely to find something
	// 1 - likely
	// 2 - very likely
	RecommendationLevel int `json:"recommendation_level"`
}

type Finder interface {
	Init()
	GetSearchableColumns() []SearchableColumn
	GenerateQuery(colkey string, params map[string][]string) (squirrel.SelectBuilder, error)
}

type FormatedColumns struct {
	Tags        []string
	PrimaryType string
	Formats     map[string]string
}

func SeekForFormatColumns(formats ...string) map[string]FormatedColumns {
	ret := make(map[string]FormatedColumns)

	for _, table := range GlobalContext.Tables {
		for colname, colinf := range table.columns {
			if colinf.GeneratorData != nil {
				if slices.Contains(formats, colinf.GeneratorData.Format) {
					col := ret[table.tableName+":"+colinf.GeneratorData.LinkedColumn]

					/*	saving one colinf should be enough to determine
						if it is compatible with specific finder */
					col.PrimaryType = colinf.GeneratorData.PrimaryType
					col.Tags = colinf.GeneratorData.Tags

					if col.Formats == nil {
						col.Formats = make(map[string]string)
					}

					col.Formats[colinf.GeneratorData.Format] = colname

					ret[table.tableName+":"+colinf.GeneratorData.LinkedColumn] = col
				}
			}
		}
	}

	return ret
}
