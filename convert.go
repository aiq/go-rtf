package rtf

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

var specialCharacters = map[string]string{
	"par":       "\n",
	"sect":      "\n\n",
	"page":      "\n\n",
	"line":      "\n",
	"tab":       "\t",
	"emdash":    "\u2014",
	"endash":    "\u2013",
	"emspace":   "\u2003",
	"enspace":   "\u2002",
	"qmspace":   "\u2005",
	"bullet":    "\u2022",
	"lquote":    "\u2018",
	"rquote":    "\u2019",
	"ldblquote": "\u201C",
	"rdblquote": "\u201D",
}

var charmaps = map[string]*charmap.Charmap{
	"437": charmap.CodePage437,
	//	"708":  nil,
	//	"709":  nil,
	//	"710":  nil,
	//	"711":  nil,
	//	"720":  nil,
	//	"819":  nil,
	"850": charmap.CodePage850,
	"852": charmap.CodePage852,
	"860": charmap.CodePage860,
	"862": charmap.CodePage862,
	"863": charmap.CodePage863,
	//	"864":  nil,
	"865": charmap.CodePage865,
	"866": charmap.CodePage866,
	//	"874":  nil,
	//	"932":  nil,
	//	"936":  nil,
	//	"949":  nil,
	//	"950":  nil,
	"1250": charmap.Windows1250,
	"1251": charmap.Windows1251,
	"1252": charmap.Windows1252,
	"1253": charmap.Windows1253,
	"1254": charmap.Windows1254,
	"1255": charmap.Windows1255,
	"1256": charmap.Windows1256,
	"1257": charmap.Windows1257,
	"1258": charmap.Windows1258,
	//	"1361": nil,
}

var rtfRegex = regexp.MustCompile(
	"(?i)" +
		`\\([a-z]{1,32})(-?\d{1,10})?[ ]?` +
		`|\\'([0-9a-f]{2})` +
		`|\\([^a-z])` +
		`|([{}])` +
		`|[\r\n]+` +
		`|(.)`)

type stackEntry struct {
	UCSkip    int
	Ignorable bool
	Actions   *Actions
}

type stackType []stackEntry

func (stack stackType) Len() int {
	return len(stack)
}

func (stack stackType) last() stackEntry {
	n := stack.Len()
	return stack[n-1]
}

func (stack stackType) Ignorable() bool {
	return stack.last().Ignorable
}

func (stack stackType) UCSkip() int {
	return stack.last().UCSkip
}

func (stack stackType) Actions() *Actions {
	return stack.last().Actions
}

func (stack stackType) SetIgnorable(v bool) {
	n := stack.Len()
	stack[n-1].Ignorable = v
}

func (stack stackType) SetUCSkip(v int) {
	n := stack.Len()
	stack[n-1].UCSkip = v
}

func newStackEntry(ucskip int, ignorable bool) stackEntry {
	return stackEntry{
		UCSkip:    ucskip,
		Ignorable: ignorable,
		Actions:   &Actions{},
	}
}

func AsActions(input string, rules RuleSet, ignore []string) (*Actions, error) {
	toIgnore := listAsSet(ignore)
	var charMap *charmap.Charmap = nil
	stack := stackType{
		newStackEntry(1, false),
	}
	curskip := 0

	matches := rtfRegex.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		word := match[1]
		arg := match[2]
		hex := match[3]
		character := match[4]
		brace := match[5]
		tchar := match[6]

		switch {
		case tchar != "":
			if curskip > 0 {
				curskip -= 1
			} else if !stack.Ignorable() {
				stack.Actions().AppendString(tchar)
			}
		case brace != "":
			curskip = 0
			if brace == "{" {
				stack = append(
					stack, newStackEntry(stack.UCSkip(), stack.Ignorable()))
			} else if brace == "}" {
				l := len(stack)
				actions := stack.Actions()
				stack = stack[:l-1]
				stack.Actions().Append(actions.Action())
			}
		case character != "":
			curskip = 0
			if character == "~" {
				if !stack.Ignorable() {
					stack.Actions().AppendString("\xA0")
				}
			} else if strings.Contains("{}\\", character) {
				if !stack.Ignorable() {
					stack.Actions().AppendString(character)
				}
			} else if character == "*" {
				stack.SetIgnorable(true)
			}
		case word != "":
			action := Action{Word: word}
			if p, e := strconv.Atoi(arg); e == nil {
				action.Para = &p
			}
			curskip = 0
			if toIgnore[word] {
				stack.SetIgnorable(true)
			} else if stack.Ignorable() {
			} else if rule, ok := rules[word]; ok {
				rule(Header{}, stack.Actions(), action)
			} else if word == "ansicpg" {
				charMap = charmaps[arg]
			} else if word == "uc" {
				i, _ := strconv.Atoi(arg)
				stack.SetUCSkip(i)
			} else if word == "u" {
				c, _ := strconv.Atoi(arg)
				if c < 0 {
					c += 0x10000
				}
				stack.Actions().AppendRune(rune(c))
				curskip = stack.UCSkip()
			}
		case hex != "":
			if curskip > 0 {
				curskip -= 1
			} else if !stack.Ignorable() {
				c, _ := strconv.ParseInt(hex, 16, 0)
				if charMap == nil {
					stack.Actions().AppendRune(rune(c))
				} else {
					stack.Actions().AppendRune(charMap.DecodeByte(byte(c)))
				}
			}
		}
	}

	return stack.Actions(), nil
}

func Convert(input string, rules RuleSet, ignore []string) (string, error) {
	actions, err := AsActions(input, rules, ignore)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	actions.Execute(&b)
	return b.String(), nil
}
