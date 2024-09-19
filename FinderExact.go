package main

import (
	"errors"
	"strings"

	"github.com/Masterminds/squirrel"
)

type ExactFinder struct {
	TargetAnalyzer   string
	CompatibleFields map[string]bool
}

func (f *ExactFinder) Init() {
	f.CompatibleFields = map[string]bool{}
	for tblname, tbl := range GlobalContext.Tables {
		for colname, colinf := range tbl.columns {
			if _, e := colinf.Analyzers["fbid_analyzer"]; e {
				f.CompatibleFields[tblname+":"+colname] = true
			}
		}
	}
}

func (f *ExactFinder) GetSearchableColumns() []SearchableColumn {
	var ret []SearchableColumn

	for fname := range f.CompatibleFields {
		var searchable_column SearchableColumn

		parts := strings.Split(fname, ":")
		tbl, col := parts[0], parts[1]
		searchable_column.Dataset = tbl
		searchable_column.Columns = []string{col}
		searchable_column.ColKey = fname
		searchable_column.RecommendationLevel = 2
		searchable_column.Features = append(searchable_column.Features, "exact_match")
		ret = append(ret, searchable_column)
	}

	return ret
}

func (f *ExactFinder) GenerateQuery(colkey string, params map[string][]string) (squirrel.SelectBuilder, error) {
	if params["value"] == nil {
		return squirrel.SelectBuilder{}, errors.New("you need to provide `value`")
	}

	parts := strings.Split(colkey, ":")

	return squirrel.Select("*").From(parts[0]).Where(squirrel.Eq{parts[1]: params["value"][0]}), nil
}
