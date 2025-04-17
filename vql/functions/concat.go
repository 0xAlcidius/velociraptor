package functions

import (
	"context"
	"fmt"

	"github.com/Velocidex/ordereddict"
	"www.velocidex.com/golang/velociraptor/acls"
	"www.velocidex.com/golang/velociraptor/vql"
	vql_subsystem "www.velocidex.com/golang/velociraptor/vql"
	"www.velocidex.com/golang/vfilter"
	"www.velocidex.com/golang/vfilter/arg_parser"
)

type ConcatFunctionArgs struct {
	String1 string `vfilter:"required,field=string1"`
	String2 string `vfilter:"required,field=string2"`
}

type ConcatFunction struct{}

func (self ConcatFunction) Info(scope vfilter.Scope, type_map *vfilter.TypeMap) *vfilter.FunctionInfo {
	return &vfilter.FunctionInfo{
		Name:     "concat",
		Doc:      "concatonates N strings together.",
		ArgType:  type_map.AddType(scope, &ConcatFunction{}),
		Metadata: vql.VQLMetadata().Permissions(acls.PREPARE_RESULTS).Build(),
		Version:  1,
	}
}

func (self ConcatFunction) Call(ctx context.Context,
	scope vfilter.Scope,
	args *ordereddict.Dict) vfilter.Any {

	fmt.Println("[CONCAT] RUNNING")

	defer vql_subsystem.RegisterMonitor("concat", args)()

	fmt.Println("[CONCAT] STILL RUNNING0")

	arg := &ConcatFunctionArgs{}
	fmt.Println("[CONCAT] STILL RUNNING1")
	err := arg_parser.ExtractArgsWithContext(ctx, scope, args, arg)
	fmt.Println("[CONCAT] STILL RUNNING2")
	if err != nil {
		fmt.Println("[CONCAT]: ERROR:", err.Error())
		scope.Log("[CONCAT]: %s", err.Error())
		return &vfilter.Null{}
	}

	fmt.Println("[CONCAT] STRING1: ", arg.String1)
	fmt.Println("[CONCAT] STRING2: ", arg.String2)

	return &vfilter.Null{}
}

func init() {
	vql_subsystem.RegisterFunction(&ConcatFunction{})
}
