package rtf

import (
	_ "embed"
	"testing"

	"github.com/google/go-cmp/cmp"
)

//go:embed convert_test_data/three-color-lines.rtf
var threeColorLines string

//go:embed convert_test_data/with-bold-and-underline.rtf
var withBoldAndUnderline string

func TestConvert(t *testing.T) {
	cases := []struct {
		input  string
		rules  RuleSet
		ignore []string
		exp    string
	}{
		{
			input:  threeColorLines,
			rules:  PlainTextRules(),
			ignore: IgnoreList(),
			exp: `This line is the default color
This line is red
This line is the default color`,
		},
		{
			input:  withBoldAndUnderline,
			rules:  HTMLRules(),
			ignore: IgnoreList(),
			exp: `<b><u>bold-underline in group</u></b><br>
<b>bold</b>-<u>underline</u> in group<br>
<b><u>bold-underline</u></b>`,
		},
	}

	for i, c := range cases {
		res, err := Convert(c.input, c.rules, c.ignore)
		if err != nil {
			t.Errorf("%v - unexpected error: %v", i, err)
			continue
		}
		diff := cmp.Diff(res, c.exp)
		if diff != "" {
			t.Errorf("%v - unexpected diff:\n%v", i, diff)
		}
	}
}
