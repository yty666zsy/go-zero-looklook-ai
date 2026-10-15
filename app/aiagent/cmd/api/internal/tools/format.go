package tools

import "time"

func intArgOr(args map[string]any, k string, def int64) int64 {
	if v, ok := intArg(args, k); ok && v > 0 {
		if k == "pageSize" && v > 10 {
			return 10
		}
		return v
	}
	return def
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func trimText(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func dateText(unix int64) string {
	if unix <= 0 {
		return "-"
	}
	return time.Unix(unix, 0).Format("2006-01-02")
}

func dateTimeText(unix int64) string {
	if unix <= 0 {
		return "-"
	}
	return time.Unix(unix, 0).Format("2006-01-02 15:04")
}
