package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
)

type PhoneFinder struct {
	FormatedColumns map[string]FormatedColumns
}

func (f *PhoneFinder) Init() {
	f.FormatedColumns = SeekForFormatColumns("numeric")
}

func (f *PhoneFinder) GetSearchableColumns() []SearchableColumn {
	var ret []SearchableColumn

	for fname, fdata := range f.FormatedColumns {
		var searchable_column SearchableColumn

		parts := strings.Split(fname, ":")
		tbl, col := parts[0], parts[1]
		searchable_column.Dataset = tbl
		searchable_column.Columns = []string{col}
		searchable_column.ColKey = fname

		if fdata.PrimaryType == "phone_number" {
			searchable_column.RecommendationLevel = 2
		}

		if _, e := fdata.Formats["numeric"]; e {
			searchable_column.Features = append(searchable_column.Features, "phone_number")
		}

		ret = append(ret, searchable_column)
	}

	return ret
}

func (f *PhoneFinder) GenerateQuery(colkey string, params map[string][]string) (squirrel.SelectBuilder, error) {
	if params["phone_number"] == nil {
		return squirrel.SelectBuilder{}, errors.New("you need to provide a `phone_number`")
	}

	phone := params["phone_number"][0]
	if !IsNumeric(phone) || phone[0] == '0' {
		return squirrel.SelectBuilder{}, errors.New("the phone number must follow the international format without +")
	}

	if _, e := f.FormatedColumns[colkey]; !e {
		return squirrel.SelectBuilder{}, fmt.Errorf("unknown colkey `%s`", colkey)
	}

	tbl := strings.Split(colkey, ":")[0]

	phones := FormatPhone(phone)

	// TODO: check numeric exist
	return squirrel.Select("*").From(tbl).Where(squirrel.Eq{
		f.FormatedColumns[colkey].Formats["numeric"]: phones}), nil
}
