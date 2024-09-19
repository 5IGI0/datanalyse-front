package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
)

type RealnameFinder struct {
	FormatedColumns map[string]FormatedColumns
}

func (f *RealnameFinder) Init() {
	f.FormatedColumns = SeekForFormatColumns("sanitized", "bidirect_sanitized")
}

func (f *RealnameFinder) GetSearchableColumns() []SearchableColumn {
	var ret []SearchableColumn

	for fname, fdata := range f.FormatedColumns {
		var searchable_column SearchableColumn

		parts := strings.Split(fname, ":")
		tbl, col := parts[0], parts[1]
		searchable_column.Dataset = tbl
		searchable_column.Columns = []string{col}
		searchable_column.ColKey = fname

		if fdata.PrimaryType == "realnames" {
			searchable_column.RecommendationLevel = 2
		} else if slices.Contains([]string{"email_login", "username"}, fdata.PrimaryType) {
			searchable_column.RecommendationLevel = 1
		}

		if _, e := fdata.Formats["sanitized"]; e {
			searchable_column.Features = append(searchable_column.Features, "single_name")
		}

		if _, e := fdata.Formats["bidirect_sanitized"]; e {
			searchable_column.Features = append(searchable_column.Features, "multi_name")
		}

		ret = append(ret, searchable_column)
	}

	return ret
}

func (f *RealnameFinder) GenerateQuery(colkey string, params map[string][]string) (squirrel.SelectBuilder, error) {
	if params["firstname"] == nil && params["lastname"] == nil {
		return squirrel.SelectBuilder{}, errors.New("you need to provide at least a firstname or a lastname")
	}

	if _, e := f.FormatedColumns[colkey]; !e {
		return squirrel.SelectBuilder{}, fmt.Errorf("unknown colkey `%s`", colkey)
	}

	names := make([]string, 0, 2)
	if params["firstname"] != nil {
		names = append(names, SanitizeText(params["firstname"][0]))
	}
	if params["lastname"] != nil {
		names = append(names, SanitizeText(params["lastname"][0]))
	}

	tbl := strings.Split(colkey, ":")[0]

	/* TODO: check if sanitized / bidirect_sanitized exist */
	if len(names) == 1 {
		/* x LIKE 'john%' OR x like '%john'*/
		return squirrel.Select("*").From(tbl).Where(
			squirrel.Or{
				squirrel.Like{
					f.FormatedColumns[colkey].Formats["sanitized"]: SQLEscapeStringLike(names[0]) + "%"},
				ReverseLike(tbl, f.FormatedColumns[colkey].Formats["sanitized"], "%"+names[0]),
			}), nil
	} else {
		/* x LIKE 'john%doe' OR x LIKE 'doe%john' */
		return squirrel.Select("*").From(tbl).Where(
			squirrel.Or{
				squirrel.Like{f.FormatedColumns[colkey].Formats["bidirect_sanitized"]: GenBidirectLike([]rune(names[0]), []rune(names[1]))},
				squirrel.Like{f.FormatedColumns[colkey].Formats["bidirect_sanitized"]: GenBidirectLike([]rune(names[1]), []rune(names[0]))}}), nil
	}
}
