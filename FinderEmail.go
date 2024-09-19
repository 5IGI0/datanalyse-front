package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
)

type EmailFinder struct {
	FormatedColumns map[string]FormatedColumns
}

func (f *EmailFinder) Init() {
	f.FormatedColumns = SeekForFormatColumns("sanitized", "bidirect_sanitized", "reverse_domain")

	for k, v := range f.FormatedColumns {
		if !slices.Contains([]string{"email_domain", "email_login"}, v.PrimaryType) {
			delete(f.FormatedColumns, k)
			continue
		}
	}
}

func (f *EmailFinder) GetSearchableColumns() []SearchableColumn {
	var ret []SearchableColumn

	for fname, fdata := range f.FormatedColumns {
		var searchable_column SearchableColumn

		parts := strings.Split(fname, ":")
		tbl, col := parts[0], parts[1]
		searchable_column.Dataset = tbl
		searchable_column.Columns = []string{col}
		searchable_column.ColKey = fname
		searchable_column.RecommendationLevel = 2

		if _, e := fdata.Formats["numeric"]; e {
			searchable_column.Features = append(searchable_column.Features, "domain", "username")
		}

		ret = append(ret, searchable_column)
	}

	return ret
}

func (f *EmailFinder) GenerateQuery(colkey string, params map[string][]string) (squirrel.SelectBuilder, error) {
	if params["domain"] == nil && params["username"] == nil {
		return squirrel.SelectBuilder{}, errors.New("you need to provide at least `domain` or `username`")
	}

	if _, e := f.FormatedColumns[colkey]; !e {
		return squirrel.SelectBuilder{}, fmt.Errorf("unknown colkey `%s`", colkey)
	}

	// TODO: check if reverse_domain / sanitized exist
	// TODO: exact match
	var conds squirrel.And

	if params["domain"] != nil {
		revdomain := ReverseDomain(params["domain"][0])
		conds = append(conds, squirrel.Like{
			f.FormatedColumns[colkey].Formats["reverse_domain"]: revdomain + "%"})
	}

	if params["username"] != nil {
		username := string(reverse_str([]rune(params["username"][0])))
		conds = append(conds, squirrel.Like{
			f.FormatedColumns[colkey].Formats["sanitized"]: username + "%"})
	}

	tbl := strings.Split(colkey, ":")[0]
	return squirrel.Select("*").From(tbl).Where(conds), nil
}
