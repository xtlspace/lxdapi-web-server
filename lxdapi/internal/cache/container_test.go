package cache

import (
	"testing"

	"lxdapi/pkg/format"
)

func TestParseCPULimit(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"4", 4},
		{"8", 8},
		{"", 1},
		{"abc", 1},
		{"0", 1},
	}
	for _, c := range cases {
		if got := parseCPULimit(c.in); got != c.want {
			t.Errorf("parseCPULimit(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	cases := []struct {
		in   uint64
		want string
	}{
		{512, "512 B"},
		{1024, "1.00 KB"},
		{1048576, "1.00 MB"},
		{1610612736, "1.50 GB"},
		{1099511627776, "1.00 TB"},
	}
	for _, c := range cases {
		if got := format.BytesUint64(c.in); got != c.want {
			t.Errorf("FormatBytesUint64(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseMemoryStringToMB(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"512MB", 512},
		{"2GB", 2048},
		{"1tb", 1024 * 1024},
		{"512KB", 0},
		{"1073741824", 1024},
		{"", 0},
		{"  2 gb ", 2048},
	}
	for _, c := range cases {
		if got := parseMemoryStringToMB(c.in); got != c.want {
			t.Errorf("parseMemoryStringToMB(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestFormatMBToString(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{0, ""},
		{512, "512MB"},
		{1023, "1023MB"},
		{1024, "1.0GB"},
		{1536, "1.5GB"},
		{2048, "2.0GB"},
	}
	for _, c := range cases {
		if got := formatMBToString(c.in); got != c.want {
			t.Errorf("formatMBToString(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}
