package utils

import "strconv"

func Int64ToStr(num int64) string {
	return strconv.FormatInt(num, 10)
}
