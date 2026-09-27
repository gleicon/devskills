package sync

import (
	"bytes"
	"errors"
	"fmt"
	"slices"

	"github.com/tailscale/hujson"
)

// OpenCode ignores disable-model-invocation, so a permission rule is what keeps
// devskills user-invoked there: denied skills drop off the model's skill list
// but still run from the Skills picker.
const skillPattern = "ds-*"

// freshConfig is the whole config when the assistant has none yet.
const freshConfig = `{
  "permission": {
    "skill": {
      "ds-*": "deny"
    }
  }
}
`

// denySkills returns the OpenCode config src with "ds-*": "deny" as the last
// rule of permission.skill — last because OpenCode applies the last matching
// rule. Comments, formatting and trailing commas in src are kept.
func denySkills(src []byte) ([]byte, error) {
	if len(bytes.TrimSpace(src)) == 0 {
		return []byte(freshConfig), nil
	}
	root, top, err := parseConfig(src)
	if err != nil {
		return nil, err
	}
	perm, err := ensureObject(top, "permission")
	if err != nil {
		return nil, err
	}
	rules, err := ensureObject(perm, "skill")
	if err != nil {
		return nil, err
	}
	if n := len(rules.Members); n > 0 && memberName(rules.Members[n-1]) == skillPattern && isDeny(rules.Members[n-1]) {
		return src, nil
	}
	for i := indexOf(rules, skillPattern); i >= 0; i = indexOf(rules, skillPattern) {
		removeMember(rules, i)
	}
	appendMember(rules, skillPattern, hujson.String("deny"))
	return root.Pack(), nil
}

// allowSkills undoes denySkills: it removes the "ds-*" rule, then the skill and
// permission objects if that leaves them empty.
func allowSkills(src []byte) ([]byte, error) {
	root, top, err := parseConfig(src)
	if err != nil {
		return nil, err
	}
	perm, rules := objectAt(&root, "/permission"), objectAt(&root, "/permission/skill")
	if rules == nil || indexOf(rules, skillPattern) < 0 {
		return src, nil
	}
	for i := indexOf(rules, skillPattern); i >= 0; i = indexOf(rules, skillPattern) {
		removeMember(rules, i)
	}
	if len(rules.Members) == 0 {
		removeMember(perm, indexOf(perm, "skill"))
	}
	if len(perm.Members) == 0 {
		removeMember(top, indexOf(top, "permission"))
	}
	return root.Pack(), nil
}

func parseConfig(src []byte) (hujson.Value, *hujson.Object, error) {
	root, err := hujson.Parse(src)
	if err != nil {
		return hujson.Value{}, nil, err
	}
	top, ok := root.Value.(*hujson.Object)
	if !ok {
		return hujson.Value{}, nil, errors.New("top level is not an object")
	}
	return root, top, nil
}

// objectAt returns the object at the JSON pointer ptr, or nil when there is
// none.
func objectAt(root *hujson.Value, ptr string) *hujson.Object {
	if v := root.Find(ptr); v != nil {
		o, _ := v.Value.(*hujson.Object)
		return o
	}
	return nil
}

// ensureObject returns obj's member name as an object, adding an empty one when
// it is missing. A string value is one rule for every pattern, so it becomes
// that object's "*" rule.
func ensureObject(obj *hujson.Object, name string) (*hujson.Object, error) {
	i := indexOf(obj, name)
	if i < 0 {
		o := &hujson.Object{}
		appendMember(obj, name, o)
		return o, nil
	}
	v := &obj.Members[i].Value
	switch t := v.Value.(type) {
	case *hujson.Object:
		return t, nil
	case hujson.Literal:
		if t.Kind() == '"' {
			o := &hujson.Object{}
			appendMember(o, "*", t)
			v.Value = o
			return o, nil
		}
	}
	return nil, fmt.Errorf("%q is neither an object nor a string", name)
}

// appendMember adds name: val as obj's last member, starting where the member
// before it starts and keeping obj's trailing-comma style.
func appendMember(obj *hujson.Object, name string, val hujson.ValueTrimmed) {
	m := hujson.ObjectMember{
		Name:  hujson.Value{Value: hujson.String(name)},
		Value: hujson.Value{BeforeExtra: hujson.Extra(" "), Value: val},
	}
	if n := len(obj.Members); n > 0 {
		last := obj.Members[n-1]
		m.Name.BeforeExtra = lineStart(last.Name.BeforeExtra)
		if last.Value.AfterExtra != nil {
			m.Value.AfterExtra = hujson.Extra{}
		}
	}
	obj.Members = append(obj.Members, m)
}

// removeMember deletes obj's member i. hujson marks a trailing comma by a
// non-nil AfterExtra on the last value, so the new last value inherits the
// removed one's mark.
func removeMember(obj *hujson.Object, i int) {
	trailing := obj.Members[len(obj.Members)-1].Value.AfterExtra != nil
	obj.Members = slices.Delete(obj.Members, i, i+1)
	if len(obj.Members) == 0 {
		return
	}
	last := &obj.Members[len(obj.Members)-1].Value
	switch {
	case trailing && last.AfterExtra == nil:
		last.AfterExtra = hujson.Extra{}
	case !trailing && last.AfterExtra != nil:
		// Concat, not append: Extra aliases the parsed input buffer.
		obj.AfterExtra = slices.Concat(last.AfterExtra, obj.AfterExtra)
		last.AfterExtra = nil
	}
}

// lineStart is the whitespace before a member that puts the next one on a line
// of its own at the same indentation, leaving behind any comments above it.
func lineStart(before hujson.Extra) hujson.Extra {
	i := bytes.LastIndexByte(before, '\n')
	if i < 0 {
		return hujson.Extra(" ")
	}
	indent := before[i+1:]
	indent = indent[:len(indent)-len(bytes.TrimLeft(indent, " \t"))]
	return slices.Concat(hujson.Extra("\n"), indent)
}

func indexOf(obj *hujson.Object, name string) int {
	return slices.IndexFunc(obj.Members, func(m hujson.ObjectMember) bool { return memberName(m) == name })
}

func memberName(m hujson.ObjectMember) string { return m.Name.Value.(hujson.Literal).String() }

func isDeny(m hujson.ObjectMember) bool {
	lit, ok := m.Value.Value.(hujson.Literal)
	return ok && lit.Kind() == '"' && lit.String() == "deny"
}
