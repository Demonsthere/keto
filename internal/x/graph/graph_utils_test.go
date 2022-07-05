package graph

import (
	"context"
	"sync"
	"testing"

	"github.com/ory/keto/internal/relationtuple"

	"github.com/stretchr/testify/assert"
)

func TestEngineUtilsProvider_CheckVisited(t *testing.T) {
	t.Run("case=finds cycle", func(t *testing.T) {
		linkedList := []relationtuple.SubjectSet{{
			Namespace: "default",
			Object:    "A",
			Relation:  "connected",
		}, {
			Namespace: "default",
			Object:    "B",
			Relation:  "connected",
		}, {
			Namespace: "default",
			Object:    "C",
			Relation:  "connected",
		}, {
			Namespace: "default",
			Object:    "B",
			Relation:  "connected",
		}, {
			Namespace: "default",
			Object:    "D",
			Relation:  "connected",
		}}

		ctx := context.Background()
		var isThereACycle bool
		for i := range linkedList {
			ctx, isThereACycle = CheckAndAddVisited(ctx, &linkedList[i])
			if isThereACycle {
				break
			}
		}

		assert.Equal(t, isThereACycle, true)
	})

	t.Run("case=ignores if no cycle", func(t *testing.T) {
		list := []relationtuple.SubjectSet{{
			Namespace: "default",
			Object:    "A",
			Relation:  "connected",
		}, {
			Namespace: "default",
			Object:    "B",
			Relation:  "connected",
		}, {
			Namespace: "default",
			Object:    "C",
			Relation:  "connected",
		}, {
			Namespace: "default",
			Object:    "D",
			Relation:  "connected",
		}, {
			Namespace: "default",
			Object:    "E",
			Relation:  "connected",
		}}

		ctx := context.Background()
		var isThereACycle bool
		for i := range list {
			ctx, isThereACycle = CheckAndAddVisited(ctx, &list[i])
			if isThereACycle {
				break
			}
		}

		assert.Equal(t, isThereACycle, false)
	})

	t.Run("case=no race condition during adding", func(t *testing.T) {
		// we repeat this test a few times to ensure we don't have a race condition
		// the race detector alone was not able to catch it
		for i := 0; i < 500; i++ {
			subject := &relationtuple.SubjectSet{
				Namespace: "default",
				Object:    "racy",
				Relation:  "connected",
			}

			ctx, _ := CheckAndAddVisited(context.Background(), &relationtuple.SubjectSet{Object: "just to setup the context"})
			var wg sync.WaitGroup
			var aCycle, bCycle bool
			var aCtx, bCtx context.Context

			wg.Add(2)
			go func() {
				aCtx, aCycle = CheckAndAddVisited(ctx, subject)
				wg.Done()
			}()
			go func() {
				bCtx, bCycle = CheckAndAddVisited(ctx, subject)
				wg.Done()
			}()

			wg.Wait()
			// one should be true, and one false
			assert.False(t, aCycle && bCycle)
			assert.True(t, aCycle || bCycle)
			assert.Equal(t, aCtx, bCtx)
		}
	})
}
