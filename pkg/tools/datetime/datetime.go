package datetime

import "time"

// 获取当前时间
func Now() time.Time {
	return time.Now()
}

// 格式化时间
func FormatDate(t time.Time, layout string) string {
	return t.Format(layout)
}

// 解析时间
func ParseDate(value, layout string) (time.Time, error) {
	return time.Parse(layout, value)
}

// 日期加减天数
func AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

// 计算两个日期之间的天数差
func DaysBetween(start, end time.Time) int {
	return int(end.Sub(start).Hours() / 24)
}

// 获取当前日期
func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// 判断是否为闰年
func IsLeapYear(year int) bool {
	return (year%4 == 0 && year%100 != 0) || (year%400 == 0)
}

// 获取指定日期的开始时间
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// 获取指定日期的结束时间
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

// GetWeekDay 获取星期几
func GetWeekDay(t time.Time) time.Weekday {
	return t.Weekday()
}

// GetQuarter 获取季度
func GetQuarter(t time.Time) int {
	month := t.Month()
	return (int(month) + 2) / 3
}

// AddMonths 增加月份
func AddMonths(t time.Time, months int) time.Time {
	return t.AddDate(0, months, 0)
}

// GetWeekOfYear 获取年内第几周
func GetWeekOfYear(t time.Time) int {
	_, week := t.ISOWeek()
	return week
}
