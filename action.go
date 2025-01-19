package rtf

import "bytes"

type Write func(b *bytes.Buffer)

type Action struct {
	Word  string
	Para  *int
	Write Write
	Close Write
}

type Actions struct {
	list []Action
}

func (as Actions) Execute(b *bytes.Buffer) {
	for _, a := range as.list {
		a.Write(b)
	}
	n := len(as.list)
	for i := n - 1; i >= 0; i-- {
		close := as.list[i].Close
		if close != nil {
			close(b)
		}
	}
}

func (as *Actions) Action() Action {
	return Action{
		Write: func(b *bytes.Buffer) {
			as.Execute(b)
		},
	}
}

func (as *Actions) Append(a Action) {
	as.list = append(as.list, a)
}

func (as *Actions) AppendString(s string) {
	as.Append(Action{
		Write: func(b *bytes.Buffer) {
			b.WriteString(s)
		},
	})
}

func (as *Actions) AppendRune(r rune) {
	as.Append(Action{
		Write: func(b *bytes.Buffer) {
			b.WriteRune(r)
		},
	})
}

type Rule func(header Header, actions *Actions, act Action) error

func As(val string) Rule {
	return func(header Header, actions *Actions, act Action) error {
		act.Write = func(b *bytes.Buffer) {
			b.WriteString(val)
		}
		actions.Append(act)
		return nil
	}
}

func Cover(head, tail string) Rule {
	return func(header Header, actions *Actions, act Action) error {
		act.Write = func(b *bytes.Buffer) {
			b.WriteString(head)
		}
		act.Close = func(b *bytes.Buffer) {
			b.WriteString(tail)
		}
		actions.Append(act)
		return nil
	}
}

func Toggle(head, tail string) Rule {
	findPref := func(actions *Actions, word string) int {
		n := len(actions.list)
		for i := n - 1; i >= 0; i-- {
			if actions.list[i].Word == word {
				return i
			}
		}
		return -1
	}
	return func(header Header, actions *Actions, act Action) error {
		if act.Para != nil && *act.Para == 0 {
			if i := findPref(actions, act.Word); i != -1 {
				actions.list[i].Close = nil
			}
			act.Write = func(b *bytes.Buffer) {
				b.WriteString(tail)
			}
		} else {
			act.Write = func(b *bytes.Buffer) {
				b.WriteString(head)
			}
			act.Close = func(b *bytes.Buffer) {
				b.WriteString(tail)
			}
		}
		actions.Append(act)
		return nil
	}
}

type RuleSet map[string]Rule
