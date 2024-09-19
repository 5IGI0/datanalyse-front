package main

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

func ApiSearch(r *http.Request) (any, error) {
	vars := mux.Vars(r)
	finder_str := vars["finder"]
	colkey := vars["colkey"]

	finder := GlobalContext.Finders[finder_str]
	if finder == nil {
		return nil, errors.New("unknown finder")
	}

	r.ParseForm()
	query, err := finder.GenerateQuery(colkey, r.Form)
	if err != nil {
		return nil, err
	}

	q, v := query.Limit(50).MustSql()

	//return []any{q, v}, nil

	rows, err := GlobalContext.Database.Queryx(q, v...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ret []map[string]any
	for rows.Next() {
		var tmp = make(map[string]any)
		AssertError(rows.MapScan(tmp))

		// cast it to string so json.Marshaler won't encode it in base64
		for k, v := range tmp {
			switch vv := v.(type) {
			case []byte:
				tmp[k] = string(vv)
			}
		}

		ret = append(ret, tmp)
	}

	return ret, nil
}
