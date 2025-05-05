package vmfs

import (
	"fmt"

	"www.velocidex.com/golang/velociraptor/accessors"
	"www.velocidex.com/golang/vfilter"
)

// This is an accessor which represents a VMFS filesystem

type VMFSFileInfo struct {
	info       *VMFSFileInfo
	_full_path *accessors.OSPath
}

type VMFSFileSystemAccessor struct {
	scope vfilter.Scope
}

func (self VMFSFileSystemAccessor) New(scope vfilter.Scope) (accessors.FileSystemAccessor, error) {
	return &VMFSFileSystemAccessor{
		scope: scope,
	}, nil
}

func (self VMFSFileSystemAccessor) Lstat(filename string) (
	accessors.FileInfo, error) {

	return nil, nil
}

func (self VMFSFileSystemAccessor) LstatWithOSPath(filename *accessors.OSPath) (
	accessors.FileInfo, error) {

	return nil, nil
}

func (self VMFSFileSystemAccessor) Open(filename string) (
	accessors.ReadSeekCloser, error) {

	return nil, nil
}

func (self VMFSFileSystemAccessor) OpenWithOSPath(filename *accessors.OSPath) (
	accessors.ReadSeekCloser, error) {

	return nil, nil
}

func (self VMFSFileSystemAccessor) ParsePath(path string) (
	*accessors.OSPath, error) {
	osPath, err := accessors.NewLinuxOSPath(path)
	if err != nil {
		return nil, err
	}

	fmt.Println("Parsed path:", osPath)

	delegateAccessor := osPath.DelegateAccessor
	delegatePath := osPath.DelegatePath

	fmt.Println("Delegate accessor func: "+delegateAccessor()+", Delegate accessor: ", delegateAccessor)
	fmt.Println("Delegate path func: "+delegatePath()+", Delegate path: ", delegatePath)

	return osPath, nil
}

func (self VMFSFileSystemAccessor) ReadDir(filename string) (
	[]accessors.FileInfo, error) {

	return nil, nil
}

func (self VMFSFileSystemAccessor) ReadDirWithOSPath(
	filename *accessors.OSPath) (
	result []accessors.FileInfo, err error) {

	return nil, nil
}

func init() {
	accessors.Register("raw_vmfs", &VMFSFileSystemAccessor{},
		`Access a VMFS filesystem.`)
}
