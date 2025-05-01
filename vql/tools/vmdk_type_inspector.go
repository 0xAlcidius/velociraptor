package tools

import (
	"context"
	"fmt"
	"os"

	"github.com/Velocidex/ordereddict"
	"www.velocidex.com/golang/velociraptor/accessors"
	"www.velocidex.com/golang/velociraptor/acls"
	"www.velocidex.com/golang/velociraptor/vql"
	vql_subsystem "www.velocidex.com/golang/velociraptor/vql"
	vfilter "www.velocidex.com/golang/vfilter"
	"www.velocidex.com/golang/vfilter/arg_parser"
)

type VmdkTypeInspectorPluginArgs struct {
	Filenames []*accessors.OSPath `vfilter:"required,field=filename,doc=A list of log files to parse."`
	Accessor  string              `vfilter:"optional,field=accessor,doc=The accessor to use."`
}

type VmdkTypeInspectorPlugin struct{}

func (self VmdkTypeInspectorPlugin) Info(scope vfilter.Scope, type_map *vfilter.TypeMap) *vfilter.PluginInfo {
	return &vfilter.PluginInfo{
		Name:     "vmdk_type_inspector",
		Doc:      "identifies the type of a vmdk file.",
		ArgType:  type_map.AddType(scope, &VmdkTypeInspectorPlugin{}),
		Metadata: vql.VQLMetadata().Permissions(acls.PREPARE_RESULTS).Build(),
		Version:  1,
	}
}

func (self VmdkTypeInspectorPlugin) Call(ctx context.Context,
	scope vfilter.Scope,
	args *ordereddict.Dict) <-chan vfilter.Row {
	output_chan := make(chan vfilter.Row)

	fmt.Println("[VMDK_TYPE_INSPECTOR] VMDK Parser called")

	go func() {
		defer close(output_chan)
		defer vql_subsystem.RegisterMonitor("vmdk_type_inspector", args)()

		arg := &VmdkTypeInspectorPluginArgs{}
		err := arg_parser.ExtractArgsWithContext(ctx, scope, args, arg)
		if err != nil {
			fmt.Println("[VMDK_TYPE_INSPECTOR] Error extracting args: ", err.Error())
			scope.Log("[CONCAT]: %s", err.Error())
			return
		}

		fmt.Println("[VMDK_TYPE_INSPECTOR] Path: ", arg.Filenames)

		err = vql_subsystem.CheckFilesystemAccess(scope, arg.Accessor)
		if err != nil {
			fmt.Println("[VMDK_TYPE_INSPECTOR] Error checking filesystem access: ", err.Error())
			return
		}

		fmt.Println("[VMDK_TYPE_INSPECTOR] Filesystem access checked")
		fmt.Println("[VMDK_TYPE_INSPECTOR] Current System: ", os.Getenv("OS"))

		accessor, err := accessors.GetAccessor(arg.Accessor, scope)
		if err != nil {
			fmt.Println("[VMDK_TYPE_INSPECTOR] Error getting accessor: ", err.Error())
			return
		}

		fmt.Println("[VMDK_TYPE_INSPECTOR] Accessor: ", accessor)

		for _, filename := range arg.Filenames {
			file := filename
			fd, err := accessor.OpenWithOSPath(file)
			if err != nil {
				fmt.Println("[VMDK_TYPE_INSPECTOR] Error opening file: ", err.Error())
				return
			}
			defer fd.Close()

			buffer := make([]byte, 512)
			bytes_read, err := fd.Read(buffer)
			if err != nil {
				fmt.Println("[VMDK_TYPE_INSPECTOR] Error reading file: ", err.Error())
				return
			}
			if bytes_read < 512 {
				fmt.Println("[VMDK_TYPE_INSPECTOR] Error: not enough bytes read")
				return
			}
			fmt.Println("[VMDK_TYPE_INSPECTOR] Buffer: ", buffer)
			fmt.Println("[VMDK_TYPE_INSPECTOR] Bytes read: ", bytes_read)
		}
	}()
	return output_chan
}

func init() {
	vql_subsystem.RegisterPlugin(&VmdkTypeInspectorPlugin{})
}
