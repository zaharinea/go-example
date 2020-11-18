package handler

import "strconv"

// Offset returns the starting number of result for pagination
func Offset(offset string, defaultVal int64) int64 {
	offsetInt, err := strconv.ParseInt(offset, 10, 64)
	if err != nil {
		offsetInt = defaultVal
	}
	if offsetInt < 0 {
		offsetInt = defaultVal
	}
	return offsetInt
}

// Limit returns the number of result for pagination
func Limit(limit string, defaultVal int64) int64 {
	limitInt, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		limitInt = defaultVal
	}
	if limitInt < 1 {
		limitInt = defaultVal
	}
	return limitInt
}
