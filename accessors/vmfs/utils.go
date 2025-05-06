package vmfs

import (
	"fmt"

	"www.velocidex.com/golang/velociraptor/accessors"
	vql_subsystem "www.velocidex.com/golang/velociraptor/vql"
	"www.velocidex.com/golang/velociraptor/vql/readers"
	"www.velocidex.com/golang/vfilter"
)

func GetVMFSContext(scope vfilter.Scope, device, fullpath *accessors.OSPath, accessor string) (
	result *VMFSContext, err error) {

	fmt.Println("[VMFS ACCESSOR] GetVMFSContext")
	if device == nil {
		device, err = fullpath.Delegate(scope)
		fmt.Println("[VMFS ACCESSOR] GetVMFSContext device", device)
		if err != nil {
			fmt.Println("[VMFS ACCESSOR] GetVMFSContext device error", err)
			return nil, err
		}
		accessor = fullpath.DelegateAccessor()
	}

	reader, err := readers.NewAccessorReader(scope, accessor, device, 0x1000)

	return GetVMFSCache(scope, device, accessor)
}

func GetVMFSCache(scope vfilter.Scope, device *accessors.OSPath, accessor string) (*VMFSContext, error) {
	key := "vmfs_cache" + device.String() + accessor

	cache_ctx, ok := vql_subsystem.CacheGet(scope, key).(*VMFSContext)
	if !ok {
		fmt.Println("[VMFS ACCESSOR] GetVMFSCache was not ok cache_ctx", cache_ctx)
	}

	return nil, nil
}
