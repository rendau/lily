package api

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type SortParsSt struct {
	Pars []SortParSt
}

type SortParSt struct {
	Column string
	Desc   bool
}

func ExtractPaginationPars(pars *url.Values) (offset uint64, limit uint64, page uint64) {
	var err error
	qPar := pars.Get("page_size")
	if qPar != "" {
		limit, err = strconv.ParseUint(qPar, 10, 64)
		if err != nil {
			limit = 0
		}
	}
	if limit == 0 {
		limit = 30
	}
	qPar = pars.Get("page")
	if qPar != "" {
		page, err = strconv.ParseUint(qPar, 10, 64)
		if err != nil {
			page = 0
		}
	}
	if page == 0 {
		page = 1
	}
	offset = (page - 1) * limit
	return
}

func ExtractSortPars(pars *url.Values, allowedColumns ...string) *SortParsSt {
	var result SortParsSt
	sort := pars.Get("sort")
	var par SortParSt
	for _, item := range strings.Split(sort, ",") {
		par = SortParSt{}
		if len(item) > 0 {
			if item[0] == '-' {
				par.Desc = true
				item = item[1:]
			}
			if len(item) > 0 {
				for _, ac := range allowedColumns {
					if ac == item {
						par.Column = item
						result.Pars = append(result.Pars, par)
						break
					}
				}
			}
		}
	}
	return &result
}

func PaginatedSortedResponse(data string, page_size, page, total uint64, sortPars *SortParsSt) string {
	var sp string
	for _, p := range sortPars.Pars {
		if sp != "" {
			sp += ","
		}
		if p.Desc {
			sp += "-"
		}
		sp += p.Column
	}
	return fmt.Sprintf(`{"page_size":%d,"page":%d,"total_count":%d,"sort":%q`, page_size, page, total, sp) +
		`,"results":` + data + `}`
}

func PaginatedResponse(data string, page_size, page, total uint64) string {
	return fmt.Sprintf(`{"page_size":%d,"page":%d,"total_count":%d`, page_size, page, total) +
		`,"results":` + data + `}`
}
