package common

import (
	"context"
	"fmt"

	"github.com/Velocidex/ordereddict"
	"www.velocidex.com/golang/velociraptor/acls"
	"www.velocidex.com/golang/velociraptor/vql"
	vql_subsystem "www.velocidex.com/golang/velociraptor/vql"
	vfilter "www.velocidex.com/golang/vfilter"
	"www.velocidex.com/golang/vfilter/arg_parser"
)

type ConcatPluginArgs struct {
	String1 string `vfilter:"required,field=string1"`
	String2 string `vfilter:"required,field=string2"`
}

type ConcatPlugin struct{}

func (self ConcatPlugin) Info(scope vfilter.Scope, type_map *vfilter.TypeMap) *vfilter.PluginInfo {
	return &vfilter.PluginInfo{
		Name:     "concat",
		Doc:      "concatonates N strings together.",
		ArgType:  type_map.AddType(scope, &ConcatPlugin{}),
		Metadata: vql.VQLMetadata().Permissions(acls.PREPARE_RESULTS).Build(),
		Version:  1,
	}
}

func (self ConcatPlugin) Call(ctx context.Context,
	scope vfilter.Scope,
	args *ordereddict.Dict) <-chan vfilter.Row {
	output_chan := make(chan vfilter.Row)

	go func() {
		defer close(output_chan)
		defer vql_subsystem.RegisterMonitor("concat", args)()

		arg := &ConcatPluginArgs{}
		err := arg_parser.ExtractArgsWithContext(ctx, scope, args, arg)
		if err != nil {
			fmt.Println("[CONCAT]: ERROR:", err.Error())
			scope.Log("[CONCAT]: %s", err.Error())
			return
		}

		fmt.Println("[CONCAT] STRING1: ", arg.String1)
		fmt.Println("[CONCAT] STRING2: ", arg.String2)

		newString := arg.String1 + arg.String2
		output_chan <- newString
	}()

	return output_chan

}

func init() {
	vql_subsystem.RegisterPlugin(&ConcatPlugin{})
}
