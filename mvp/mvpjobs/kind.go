package mvpjobs

import (
	"fmt"
	"time"

	"github.com/andreyvit/buddyd/mvp/backoff"
	"github.com/andreyvit/buddyd/mvp/mvprpc"
)

type Kind struct {
	schema         *Schema
	Name           string
	Behavior       Behavior
	Persistence    Persistence
	Method         *mvprpc.Method
	Set            string
	RepeatInterval time.Duration
	Backoff        backoff.Backoff
}

func (k *Kind) IsCron() bool {
	return k.Behavior == Cron
}

func (k *Kind) Retry(bo backoff.Backoff) *Kind {
	k.Backoff = bo
	return k
}

func (k *Kind) AllowNames() bool {
	return !k.IsCron()
}

func (schema *Schema) Kinds() []*Kind {
	kinds := make([]*Kind, 0, len(schema.kinds))
	for _, kind := range schema.kinds {
		kinds = append(kinds, kind)
	}
	return kinds
}

func (schema *Schema) Define(kindName string, in any, behavior Behavior, opts ...any) *Kind {
	kind := &Kind{
		schema:   schema,
		Behavior: behavior,
		Name:     kindName,
		Method:   schema.api.Method("Job"+kindName, in, nil),
	}
	if schema.kinds[kindName] != nil {
		panic(fmt.Errorf("duplicate kind %q", kindName))
	}
	schema.kinds[kindName] = kind
	for _, opt := range opts {
		switch opt := opt.(type) {
		case Persistence:
			kind.Persistence = opt
		default:
			panic(fmt.Errorf("%s: unknown options %T %v", kindName, opt, opt))
		}
	}
	// for _, tag := range strings.Fields(tags) {
	// 	schema.byTag[tag] = append(schema.byTag[tag], kind)
	// }
	return kind
}

func (schema *Schema) Cron(kindName string, interval time.Duration, opts ...any) *Kind {
	kind := schema.Define(kindName, &NoParams{}, Cron, opts...)
	kind.RepeatInterval = interval
	return kind
}
