package ketoapi

import (
	"fmt"
	"github.com/ory/herodot"
	"github.com/pkg/errors"
	"strings"
)

var ErrMalformedInput = herodot.ErrBadRequest.WithError("malformed string input")

func (r *RelationTuple) String() string {
	sb := strings.Builder{}
	sb.WriteString(r.Namespace)
	sb.WriteRune(':')
	sb.WriteString(r.Object)
	sb.WriteRune('#')
	sb.WriteString(r.Relation)
	sb.WriteRune('@')

	if r.SubjectID != nil {
		sb.WriteString(*r.SubjectID)
	} else if r.SubjectSet != nil {
		sb.WriteRune('(')
		sb.WriteString(r.SubjectSet.Namespace)
		sb.WriteRune(':')
		sb.WriteString(r.SubjectSet.Object)
		sb.WriteRune('#')
		sb.WriteString(r.SubjectSet.Relation)
		sb.WriteRune(')')
	} else {
		sb.WriteString("<ERROR: no subject>")
	}
	return sb.String()
}

func (r *RelationTuple) FromString(s string) (*RelationTuple, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return nil, errors.WithStack(ErrMalformedInput.WithDebug("expected input to contain ':'"))
	}
	r.Namespace = parts[0]

	parts = strings.SplitN(parts[1], "#", 2)
	if len(parts) != 2 {
		return nil, errors.WithStack(ErrMalformedInput.WithDebug("expected input to contain '#'"))
	}
	r.Object = parts[0]

	parts = strings.SplitN(parts[1], "@", 2)
	if len(parts) != 2 {
		return nil, errors.WithStack(ErrMalformedInput.WithDebug("expected input to contain '@'"))
	}
	r.Relation = parts[0]

	// remove optional brackets around the subject set
	sub := strings.Trim(parts[1], "()")
	if strings.Contains(sub, "#") {
		subSet, err := (&SubjectSet{}).FromString(sub)
		if err != nil {
			return nil, err
		}
		r.SubjectSet = subSet
	} else {
		r.SubjectID = &sub
	}

	return r, nil
}

func (s *SubjectSet) String() string {
	return fmt.Sprintf("%s:%s#%s", s.Namespace, s.Object, s.Relation)
}

func (s *SubjectSet) FromString(str string) (*SubjectSet, error) {
	parts := strings.Split(str, "#")
	if len(parts) != 2 {
		return nil, errors.WithStack(ErrMalformedInput.WithDebug("expected subject set to contain '#'"))
	}

	innerParts := strings.Split(parts[0], ":")
	if len(innerParts) != 2 {
		return nil, errors.WithStack(ErrMalformedInput.WithDebug("expected subject set to contain ':'"))
	}

	return &SubjectSet{
		Namespace: innerParts[0],
		Object:    innerParts[1],
		Relation:  parts[1],
	}, nil
}

func (t *ExpandTree) String() string {
	if t == nil {
		return ""
	}

	sub := "<!--no subject-->"
	switch {
	case t.SubjectID != nil:
		sub = *t.SubjectID
	case t.SubjectSet != nil:
		sub = t.SubjectSet.String()
	}

	if t.Type == Leaf {
		return fmt.Sprintf("☘ %s️", sub)
	}

	children := make([]string, len(t.Children))
	for i, c := range t.Children {
		children[i] = strings.Join(strings.Split(c.String(), "\n"), "\n│  ")
	}

	return fmt.Sprintf("∪ %s\n├─ %s", sub, strings.Join(children, "\n├─ "))
}