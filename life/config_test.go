package life

import (
	"testing"

	"github.com/go-ini/ini"
)

var config1 = []byte(`
[Schedule.1]
On = 12:00
Off = 12:30

[Schedule.2]
Days = Mon, Tue, Wed, Thu, Fri
On = 08:00
Off = 08:45

[Schedule.3]
Days = Mon, Tue, Wed, Thu, Fri
On = 17:15
Off = 23:00

[Schedule.4]
Days = Sat, Sun
On = 08:00
Off = 23:00

[Schedule.5]
Days = Sat
On = 02:00
Off = 06:00

[Schedule.6]
Days = Sun
On = 06:00
Off = 23:00
`)

func TestTimeCompare(t *testing.T) {
	ts := [...]Time{
		toTime("12:00", false),
		toTime("12:00", false),
		toTime("12:30", false),
		toTime("12:30", false),
		toTime("16:00", false),
		toTime("08:30", false),
		toTime("08:00", false),
	}

	want := [...]int{0, -1, 0, -1, 1, 1}

	for i, w := range want {
		tm1 := ts[i]
		tm2 := ts[i+1]
		c := tm1.Compare(tm2)
		if c != w {
			t.Errorf("compare(%s, %s) = %d; want %d", tm1, tm2, c, w)
		}
	}
}

func TestConfigLoadSchedules(t *testing.T) {
	f, _ := ini.Load(config1)

	c := NewConfig()
	c.loadSchedules(f)

	weekday := [6]Time{
		toTime("08:00", true),
		toTime("08:45", false),
		toTime("12:00", true),
		toTime("12:30", false),
		toTime("17:15", true),
		toTime("23:00", false),
	}
	want := map[string][6]Time{
		"Mon": weekday,
		"Tue": weekday,
		"Wed": weekday,
		"Thu": weekday,
		"Fri": weekday,
		"Sat": {
			toTime("02:00", true),
			toTime("06:00", false),
			toTime("08:00", true),
			toTime("12:00", true),
			toTime("12:30", false),
			toTime("23:00", false),
		},
		"Sun": {
			toTime("06:00", true),
			toTime("08:00", true),
			toTime("12:00", true),
			toTime("12:30", false),
			toTime("23:00", false),
			toTime("23:00", false),
		},
	}

	for d, ts := range want {
		s, ok := c.schedules[d]
		if !ok {
			t.Errorf("want schedules[%s]", d)
		}
		for i, tm := range ts {
			c := tm.Compare(s[i])
			if c != 0 {
				t.Errorf("compare(%s, %s) = %d; want 0", tm, s[i], c)
			}
		}
	}
}

func TestConfigGetScheduleState(t *testing.T) {
	f, _ := ini.Load(config1)

	c := NewConfig()
	c.loadSchedules(f)

	daytimes := [...][2]string{
		{"Mon", "08:30"},
		{"Tue", "13:00"},
		{"Wed", "17:00"},
		{"Wed", "17:30"},
		{"Sat", "07:30"},
		{"Sat", "12:15"},
		{"Sun", "23:15"},
	}
	states := [...][2]bool{
		{false, true},
		{false, false},
		{true, false},
		{true, true},
		{true, false},
		{false, true},
		{true, false},
	}

	for i, dt := range daytimes {
		s := c.GetScheduleState(dt[0], dt[1], states[i][0])
		if s != states[i][1] {
			t.Errorf("state = %t; want %t", s, states[i][1])
		}
	}
}
