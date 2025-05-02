package tools

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Velocidex/ordereddict"
	"www.velocidex.com/golang/velociraptor/accessors"
	"www.velocidex.com/golang/velociraptor/acls"
	"www.velocidex.com/golang/velociraptor/vql"
	vql_subsystem "www.velocidex.com/golang/velociraptor/vql"
	vfilter "www.velocidex.com/golang/vfilter"
	"www.velocidex.com/golang/vfilter/arg_parser"
)

type VmdkType string

const (
	MONOLITHICSPARSE VmdkType = "MonoliothicSparse"
	MONOLITHICFLAT   VmdkType = "monolithicFlat"
	SPLITSPARSE      VmdkType = "twoGbMaxExtentSparse"
	SPLITFLAT        VmdkType = "twoGbMaxExtentFlat"
	VMFS             VmdkType = "vmfs"
	UNKNOWN          VmdkType = "unknown"
)

const (
	sectorSize = 512
)

var (
	vmdkType = UNKNOWN
)

type VmdkTypeInspectorPluginArgs struct {
	Filename *accessors.OSPath `vfilter:"required,field=filename,doc=A list of log files to parse."`
	Accessor string            `vfilter:"optional,field=accessor,doc=The accessor to use."`
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
			scope.Log("[VMDK_TYPE_INSPECTOR]: %s", err.Error())
			return
		}

		fmt.Println("[VMDK_TYPE_INSPECTOR] Path: ", arg.Filename)

		err = vql_subsystem.CheckFilesystemAccess(scope, arg.Accessor)
		if err != nil {
			fmt.Println("[VMDK_TYPE_INSPECTOR] Error checking filesystem access: ", err.Error())
			return
		}

		fmt.Println("[VMDK_TYPE_INSPECTOR] Filesystem access checked")

		accessor, err := accessors.GetAccessor(arg.Accessor, scope)
		if err != nil {
			fmt.Println("[VMDK_TYPE_INSPECTOR] Error getting accessor: ", err.Error())
			return
		}

		file := arg.Filename
		fd, err := accessor.OpenWithOSPath(file)
		if err != nil {
			fmt.Println("[VMDK_TYPE_INSPECTOR] Error opening file: ", err.Error())
			return
		}
		defer fd.Close()

		contents := make([]byte, sectorSize)
		bytesRead := 0
		for bytesRead < sectorSize {
			n, err := fd.Read(contents[bytesRead:])
			if err != nil {
				if err == io.EOF {
					break
				} else {
					fmt.Errorf("failed to read file: %w", err)
				}

			}
			if n == 0 {
				break
			}
			bytesRead += n
		}

		if err != nil {
			fmt.Println("[VMDK_TYPE_INSPECTOR] Error reading file: ", err.Error())
			return
		}

		if string(contents[:4]) == "KDMV" {
			vmdkType = MONOLITHICSPARSE
		} else if string(contents[:21]) == "# Disk DescriptorFile" {
			scanner := bufio.NewScanner(bytes.NewReader(contents))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(line, "createType") {
					if strings.Contains(line, string(MONOLITHICFLAT)) {
						vmdkType = MONOLITHICFLAT
					} else if strings.Contains(line, string(SPLITSPARSE)) {
						vmdkType = SPLITSPARSE
					} else if strings.Contains(line, string(SPLITFLAT)) {
						vmdkType = SPLITFLAT
					} else if strings.Contains(line, string(VMFS)) {
						vmdkType = VMFS
					}
					break
				}
			}
		} else {
			fmt.Println("[VMDK_TYPE_INSPECTOR] Not a supported VMDK file")
		}

		output_chan <- vfilter.Row(map[string]interface{}{
			"filename": arg.Filename.String(),
			"type":     string(vmdkType),
		})

	}()
	return output_chan
}

func init() {
	vql_subsystem.RegisterPlugin(&VmdkTypeInspectorPlugin{})
}
