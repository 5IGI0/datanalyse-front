package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
)

type UsernameFinder struct {
	FormatedColumns map[string]FormatedColumns
}

func (f *UsernameFinder) Init() {
	f.FormatedColumns = SeekForFormatColumns("sanitized", "bidirect_sanitized")
}

func (f *UsernameFinder) GetSearchableColumns() []SearchableColumn {
	var ret []SearchableColumn

	for fname, fdata := range f.FormatedColumns {
		var searchable_column SearchableColumn

		parts := strings.Split(fname, ":")
		tbl, col := parts[0], parts[1]
		searchable_column.Dataset = tbl
		searchable_column.Columns = []string{col}
		searchable_column.ColKey = fname

		if fdata.PrimaryType == "username" {
			searchable_column.RecommendationLevel = 2
		} else if slices.Contains(fdata.Tags, "username") {
			searchable_column.RecommendationLevel = 1
		}

		if _, e := fdata.Formats["sanitized"]; e {
			searchable_column.Features = append(searchable_column.Features, "prefix", "suffix")
		}

		if _, e := fdata.Formats["bidirect_sanitized"]; e {
			searchable_column.Features = append(searchable_column.Features, "bothfix")
		}

		ret = append(ret, searchable_column)
	}

	return ret
}

func (f *UsernameFinder) GenerateQuery(colkey string, params map[string][]string) (squirrel.SelectBuilder, error) {
	if params["prefix"] == nil && params["suffix"] == nil {
		return squirrel.SelectBuilder{}, errors.New("you need to provide at least a prefix or a suffix")
	}

	if _, e := f.FormatedColumns[colkey]; !e {
		return squirrel.SelectBuilder{}, fmt.Errorf("unknown colkey `%s`", colkey)
	}

	prefix := ""
	suffix := ""
	if params["prefix"] != nil {
		prefix = SanitizeText(params["prefix"][0])
	}
	if params["suffix"] != nil {
		suffix = SanitizeText(params["suffix"][0])
	}

	tbl := strings.Split(colkey, ":")[0]

	/* TODO: check if sanitized / bidirect_sanitized exist */
	if params["suffix"] == nil {
		return squirrel.Select("*").From(tbl).Where(
			squirrel.Like{f.FormatedColumns[colkey].Formats["sanitized"]: SQLEscapeStringLike(prefix) + "%"}), nil
	} else if params["prefix"] == nil {
		return squirrel.Select("*").From(tbl).Where(
			ReverseLike(tbl, f.FormatedColumns[colkey].Formats["sanitized"],
				"%"+suffix)), nil
	} else {
		return squirrel.Select("*").From(tbl).Where(
			squirrel.Like{f.FormatedColumns[colkey].Formats["bidirect_sanitized"]: GenBidirectLike([]rune(prefix), []rune(suffix))}), nil
	}
}
