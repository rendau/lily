package lily

import (
	"fmt"
	"net/url"
	"strconv"
)

func ApiExtractPaginationPars(pars *url.Values) (offset uint64, limit uint64, page uint64) {
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

func ApiExtractSortPars(pars *url.Values) (sortColumn string, sortDesc bool) {
	// TODO change this fun
	sortColumn = pars.Get("sort_columns")
	qPar := pars.Get("sort_order")
	sortDesc = qPar == "desc"
	return
}

func ApiPaginatedResponse(data string, page_size, page, total uint64) string {
	return fmt.Sprintf(`{"page_size":%d,"page":%d,"total_count":%d`, page_size, page, total) +
		`,"results":` + data + `}`
}

func ApiJsonListQuery(q string) string {
	return `(select coalesce(array_to_json(array_agg(row_to_json(sq.*))), '[]')
		from (` + q + `) sq)`
}

func ApiJsonListStrQuery(q string) string {
	return `select coalesce(array_to_json(array_agg(row_to_json(sq.*)))::text, '[]')
		from (` + q + `) sq`
}

func ApiJsonRowQuery(q string) string {
	return `(select row_to_json(sq.*) from (` + q + `) sq)`
}

func ApiJsonRowStrQuery(q string) string {
	return `select row_to_json(sq.*)::text from (` + q + `) sq`
}
