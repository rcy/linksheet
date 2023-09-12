package linkmap

import "testing"

func TestNewFromCSVString(t *testing.T) {
	m := NewFromCSVString("alias,expansion\nfoo,bar")
	if m.csvmap["alias"] != "expansion" {
		t.Errorf("alias broken")
	}
	if m.csvmap["foo"] != "bar" {
		t.Errorf("foo broken")
	}
	if m.csvmap["bogus"] != "" {
		t.Errorf("bogus broken")
	}
}

func TestLookup(t *testing.T) {
	m := NewFromCSVString(`
foo,https://example.com/foo
bar/1,https://example.com/barone
`)

	for _, tc := range []struct {
		alias string
		want  string
	}{
		{
			alias: "foo",
			want:  "https://example.com/foo",
		},
		{
			alias: "bar/1",
			want:  "https://example.com/barone",
		},
	} {
		t.Run(tc.alias, func(t *testing.T) {
			got := m.Lookup(tc.alias)
			if tc.want != got {
				t.Errorf("want '%s' got '%s'", tc.want, got)
			}
		})
	}
}
