package main

import "net/http"

func ApiStatus(_ *http.Request) (any, error) {
	type PublicDataset struct {
		Name        string            `json:"name"`
		Slug        string            `json:"slug"`
		Description string            `json:"description"`
		Meta        map[string]string `json:"meta"`
		RowCount    int               `json:"row_count"`
		Size        int               `json:"size"`
		Columns     []string          `json:"columns"`
	}
	var ret struct {
		Datasets map[string]PublicDataset      `json:"datasets"`
		Finders  map[string][]SearchableColumn `json:"finders"`
	}

	ret.Datasets = make(map[string]PublicDataset)
	for _, table := range GlobalContext.Tables {
		var dataset PublicDataset

		dataset.Slug = table.tableName
		dataset.Name = table.Name
		dataset.Description = table.Description
		dataset.Meta = table.Meta
		dataset.RowCount = table.TableInfo.RowCount
		dataset.Size = table.TableInfo.Size

		for name, col := range table.columns {
			if !col.IsInvisible {
				dataset.Columns = append(dataset.Columns, name)
			}
		}

		ret.Datasets[dataset.Slug] = dataset
	}

	ret.Finders = make(map[string][]SearchableColumn)
	for finder_name, finder := range GlobalContext.Finders {
		tmp := finder.GetSearchableColumns() // TODO: resolve virtcols
		for i := range tmp {
			tmp[i].Columns = ResolvVirtCols(tmp[i].Dataset, tmp[i].Columns)
		}
		ret.Finders[finder_name] = tmp
	}

	return ret, nil
}

func ResolvVirtCols(table string, collist []string) []string {
	ret := make([]string, 0, len(collist))

	for _, col := range collist {

		virtcol_map := GlobalContext.Tables[table].VirtColMap

		if cols, e := virtcol_map[col]; e {
			ret = append(ret, cols...)
		} else {
			ret = append(ret, col)
		}
	}

	return ret
}
