package starlibs

import (
	"strings"
	gotime "time"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var timeModule = &starlarkstruct.Module{
	Name: "time",
	Members: starlark.StringDict{
		"tzname":   _tzname(),
		"altzone":  _altzone(),
		"timezone": _timezone(),

		"time":      starlark.NewBuiltin("time", time),
		"mktime":    starlark.NewBuiltin("mktime", mktime),
		"ctime":     starlark.NewBuiltin("ctime", ctime),
		"asctime":   starlark.NewBuiltin("asctime", asctime),
		"gmtime":    starlark.NewBuiltin("gmtime", gmtime),
		"localtime": starlark.NewBuiltin("localtime", localtime),
		"strftime":  starlark.NewBuiltin("strftime", strftime),
		"strptime":  starlark.NewBuiltin("strptime", strptime),
	},
}

func _tzname() starlark.Value {
	now := gotime.Now()
	name1, _ := now.Zone()
	name2, _ := now.In(gotime.Local).Zone()
	return starlark.Tuple{
		starlark.String(name1),
		starlark.String(name2),
	}
}

// TODO 计算方式有可能不准确
func _altzone() starlark.Value {
	now := gotime.Now()
	_, offsetDST := now.In(gotime.Local).Zone()
	return starlark.MakeInt(offsetDST * -1)
}

func _timezone() starlark.Value {
	now := gotime.Now()
	_, offset := now.Zone()
	return starlark.MakeInt(offset * -1)
}

func _time() starlark.Float {
	microseconds := gotime.Now().UnixMicro()
	return starlark.Float(float64(microseconds) / 1000 / 1000)
}

func time(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return _time(), nil
}

func mktime(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var val Time
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &val); err != nil {
		return nil, err
	}

	t := gotime.Time(val)

	// 为了和 python 一致，需要减去偏移量
	_, offset := t.Zone()
	t = t.Add(gotime.Duration(offset) * gotime.Second)

	seconds := float64(t.UnixMicro()) / 1000 / 1000
	return starlark.Float(seconds), nil
}

func gmtime(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var seconds starlark.Float
	if args.Len() < 1 {
		seconds = _time()
	} else {
		if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &seconds); err != nil {
			return nil, err
		}
	}
	t := gotime.UnixMicro(int64(seconds * 1000 * 1000)).UTC()
	return Time(t), nil
}

func localtime(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var seconds starlark.Float
	if args.Len() < 1 {
		seconds = _time()
	} else {
		if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &seconds); err != nil {
			return nil, err
		}
	}
	t := gotime.UnixMicro(int64(seconds * 1000 * 1000))
	return Time(t), nil
}

func ctime(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var seconds starlark.Float
	if args.Len() < 1 {
		seconds = _time()
	} else {
		if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &seconds); err != nil {
			return nil, err
		}
	}
	t := gotime.UnixMicro(int64(seconds * 1000 * 1000))
	return starlark.String(t.Format(gotime.ANSIC)), nil
}

func asctime(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var t Time
	if len(args) < 1 {
		t = Time(gotime.Now())
	} else {
		if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &t); err != nil {
			return nil, err
		}
	}

	return starlark.String(gotime.Time(t).Format(gotime.ANSIC)), nil
}

func strftime(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (_ starlark.Value, err error) {
	var (
		format starlark.String
		val    Time = Time(gotime.Now())
	)
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &format, &val); err != nil {
		return nil, err
	}

	t := gotime.Time(val)
	goFormat := convertUnixFormat(format.GoString())

	return starlark.String(t.Format(goFormat)), nil
}

func strptime(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (_ starlark.Value, err error) {
	var (
		value  starlark.String
		format starlark.String = "%a %b %d %H:%M:%S %Y"
	)
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &value, &format); err != nil {
		return nil, err
	}

	goFormat := convertUnixFormat(format.GoString())

	t, err := gotime.Parse(goFormat, value.GoString())
	if err != nil {
		return nil, err
	}

	return Time(t), nil
}

type Time gotime.Time

// Attr implements starlark.HasAttrs.
func (t Time) Attr(name string) (starlark.Value, error) {
	gt := gotime.Time(t)
	switch name {
	case "tm_year":
		return starlark.MakeInt(gt.Year()), nil
	case "tm_mon":
		return starlark.MakeInt(int(gt.Month())), nil
	case "tm_mday":
		return starlark.MakeInt(gt.Day()), nil
	case "tm_hour":
		return starlark.MakeInt(gt.Hour()), nil
	case "tm_min":
		return starlark.MakeInt(gt.Minute()), nil
	case "tm_sec":
		return starlark.MakeInt(gt.Second()), nil
	case "tm_wday":
		return starlark.MakeInt(int(gt.Weekday())), nil
	case "tm_yday":
		return starlark.MakeInt(gt.YearDay()), nil
	case "tm_isdst":
		return starlark.Bool(gt.IsDST()), nil
	default:
		return starlark.None, nil
	}
}

// AttrNames implements starlark.HasAttrs.
func (t Time) AttrNames() []string {
	return []string{
		"tm_year",
		"tm_mon",
		"tm_mday",
		"tm_hour",
		"tm_min",
		"tm_sec",
		"tm_wday",
		"tm_yday",
		"tm_isdst",
	}
}

// Freeze implements starlark.Value.
func (t Time) Freeze() {}

// Hash implements starlark.Value.
func (t Time) Hash() (uint32, error) {
	return uint32(gotime.Time(t).UnixNano()) ^ uint32(int64(gotime.Time(t).UnixNano())>>32), nil
}

// String implements starlark.Value.
func (t Time) String() string {
	return gotime.Time(t).String()
}

// Truth implements starlark.Value.
func (t Time) Truth() starlark.Bool {
	return starlark.True
}

// Type implements starlark.Value.
func (t Time) Type() string {
	return "time.Time"
}

var (
	_ starlark.Value    = Time{}
	_ starlark.HasAttrs = Time{}
)

// 将 Unix 格式的时间格式化方式转为 Go 的时间格式化方式
func convertUnixFormat(format string) string {
	replacements := map[string]string{
		"%a": "Mon",                     // Weekday as locale’s abbreviated name
		"%b": "Jan",                     // Month as locale’s abbreviated name
		"%d": "02",                      // Day of the month as zero-padded decimal number
		"%H": "15",                      // Hour (24-hour clock) as zero-padded decimal number
		"%M": "04",                      // Minute as zero-padded decimal number
		"%S": "05",                      // Second as zero-padded decimal number
		"%Y": "2006",                    // Year with century as a decimal number
		"%m": "01",                      // Month as zero-padded decimal number
		"%y": "06",                      // Year without century as zero-padded decimal number
		"%I": "03",                      // Hour (12-hour clock) as zero-padded decimal number
		"%p": "PM",                      // AM or PM
		"%j": "002",                     // Day of the year as zero-padded decimal number
		"%A": "Monday",                  // Weekday as locale’s full name
		"%B": "January",                 // Month as locale’s full name
		"%c": "Mon Jan 2 15:04:05 2006", // Locale’s appropriate date and time representation
		"%x": "01/02/06",                // Locale’s appropriate date representation
		"%X": "15:04:05",                // Locale’s appropriate time representation
		"%Z": "MST",                     // Time zone name
		"%%": "%",                       // A literal '%' character
	}

	for k, v := range replacements {
		format = strings.ReplaceAll(format, k, v)
	}

	return format
}
