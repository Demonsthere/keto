package checkgroup

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ory/keto/internal/expand"
	"github.com/ory/keto/internal/relationtuple"
)

type (
	Checkgroup interface {
		Done() bool
		Add(check CheckFunc)
		SetIsMember()
		Result() Result
		CheckFunc() CheckFunc
	}

	Factory = func(ctx context.Context) Checkgroup

	CheckFunc = func(ctx context.Context, resultCh chan<- Result)
	Result    struct {
		Membership Membership
		Tree       *expand.Tree
		Err        error
	}

	Edge struct {
		Tuple relationtuple.InternalRelationTuple
		Type  expand.NodeType
	}

	Transformation int

	Membership int
)

//go:generate stringer -type Membership
const (
	MembershipUnknown Membership = iota
	IsMember
	NotMember
)

var (
	ResultIsMember  = Result{Membership: IsMember}
	ResultNotMember = Result{Membership: NotMember}
)

var DefaultFactory = NewConcurrent

func New(ctx context.Context) Checkgroup {
	return DefaultFactory(ctx)
}

func ErrorFunc(err error) CheckFunc {
	return func(_ context.Context, resultCh chan<- Result) {
		resultCh <- Result{Err: errors.WithStack(err)}
	}
}

var IsMemberFunc = func(_ context.Context, resultCh chan<- Result) {
	resultCh <- Result{Membership: IsMember}
}

var NotMemberFunc = func(_ context.Context, resultCh chan<- Result) {
	resultCh <- Result{Membership: NotMember}
}

var UnknownMemberFunc = func(_ context.Context, resultCh chan<- Result) {
	resultCh <- Result{Membership: MembershipUnknown}
}

// WithEdge adds the edge e to the result of the function.
func WithEdge(e Edge, f CheckFunc) CheckFunc {
	return func(ctx context.Context, resultCh chan<- Result) {
		childCh := make(chan Result, 1)
		go f(ctx, childCh)
		select {
		case result := <-childCh:
			if result.Tree == nil {
				result.Tree = &expand.Tree{
					Type:  expand.Leaf,
					Tuple: &e.Tuple,
				}
			} else {
				result.Tree = &expand.Tree{
					Type:     e.Type,
					Tuple:    &e.Tuple,
					Children: []*expand.Tree{result.Tree},
				}
			}
			resultCh <- result
		case <-ctx.Done():
			resultCh <- Result{Err: ctx.Err()}
		}
	}
}
