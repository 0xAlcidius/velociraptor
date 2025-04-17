package functions

import (
	"context"

	"github.com/Velocidex/ordereddict"
	vql_subsystem "www.velocidex.com/golang/velociraptor/vql"
	"www.velocidex.com/golang/vfilter"
)

// type ConcatFunctionArgs struct {
// 	AlertName string    `vfilter:"required,field=name,doc=Name of the alert."`
// 	DedupTime int64     `vfilter:"optional,field=dedup,doc=Suppress same message in this many seconds (default 7200 sec or 2 hours)."`
// 	Condition types.Any `vfilter:"options,field=condition,doc=If specified we ignore the alert unless the condition is true"`
// }

type ConcatFunction struct{}

func (self ConcatFunction) Info(scope vfilter.Scope, type_map *vfilter.TypeMap) *vfilter.FunctionInfo {
	return &vfilter.FunctionInfo{
		Name:    "concat",
		Doc:     "concatonates N strings together.",
		Version: 1,
	}
}

func (self ConcatFunction) Call(ctx context.Context,
	scope vfilter.Scope,
	args *ordereddict.Dict) vfilter.Any {

	new := vfilter.NewScope()
	for _, k := range args.Keys() {
		v, _ := args.Get(k)
		new.Add(new, v)
	}

	return new
}

func init() {
	vql_subsystem.RegisterFunction(&_ItemsFunc{})
}
