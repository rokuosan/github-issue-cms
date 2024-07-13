package config

import (
	"strings"
	"time"
)

// compile compiles a template with the given datetime.
// %Y - Year with century as a decimal number.
// %m - Month as a decimal number [01,12].
// %d - Day of the month as a decimal number [01,31].
// %H - Hour (24-hour clock) as a decimal number [00,23].
// %M - Minute as a decimal number [00,59].
// %S - Second as a decimal number [00,59].
func CompileTimeTemplate(datetime time.Time, template string) string {
	now := datetime

	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	hour := now.Format("15")
	minute := now.Format("04")
	second := now.Format("05")

	template = strings.ReplaceAll(template, "%Y", year)
	template = strings.ReplaceAll(template, "%m", month)
	template = strings.ReplaceAll(template, "%d", day)
	template = strings.ReplaceAll(template, "%H", hour)
	template = strings.ReplaceAll(template, "%M", minute)
	template = strings.ReplaceAll(template, "%S", second)

	return template
}
