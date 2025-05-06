package vmfs

import (
	"fmt"
	"runtime/debug"
	"sync"

	"www.velocidex.com/golang/velociraptor/accessors"
	"www.velocidex.com/golang/velociraptor/uploads"
	"www.velocidex.com/golang/vfilter"
)

// This is an accessor which represents a VMFS filesystem

type readAdapter struct {
	sync.Mutex

	info   accessors.FileInfo
	reader RangeReaderAt
	pos    int64
}

type VMFSFile struct {
	offset int64

	descriptor *FS3Descriptor
}

type VMFSFileSystemAccessor struct {
	scope vfilter.Scope

	accessor string
	device   *accessors.OSPath

	root *accessors.OSPath
}

func (self VMFSFileSystemAccessor) New(scope vfilter.Scope) (accessors.FileSystemAccessor, error) {

	fmt.Println("[VMFS ACCESSOR] New returns nil")
	return &VMFSFileSystemAccessor{
		scope: scope,
	}, nil
}

func (self VMFSFileSystemAccessor) Lstat(filename string) (
	accessors.FileInfo, error) {

	fmt.Println("[VMFS ACCESSOR] Lstat returns nil")
	return nil, nil
}

func (self VMFSFileSystemAccessor) LstatWithOSPath(filename *accessors.OSPath) (
	accessors.FileInfo, error) {

	fmt.Println("[VMFS ACCESSOR] LstatWithOSPath returns nil")
	return nil, nil
}

func (self VMFSFileSystemAccessor) Open(filename string) (
	accessors.ReadSeekCloser, error) {

	fmt.Println("[VMFS ACCESSOR] Open returns nil")
	return nil, nil
}

func (self *VMFSFileSystemAccessor) OpenWithOSPath(fullpath *accessors.OSPath) (res accessors.ReadSeekCloser, err error) {

	defer func() {
		r := recover()
		if r != nil {
			fmt.Printf("PANIC %v\n", r)
			debug.PrintStack()
			err, _ = r.(error)
		}
	}()

	device := self.device
	accessor := self.accessor

	if device == nil {
		device, err = fullpath.Delegate(self.scope)
		if err != nil {
			return nil, err
		}
		accessor = fullpath.DelegateAccessor()
	}

	fmt.Println("[VMFS ACCESSOR] OpenWithOSPath accessor:", accessor)
	fmt.Println("[VMFS ACCESSOR] OpenWithOSPath device:", device)
	fmt.Println("[VMFS ACCESSOR] OpenWithOSPath fullpath:", fullpath)

	fmt.Println("[VMFS ACCESSOR] OpenWithOSPath returns nil")
	return nil, nil
}

func (self VMFSFileSystemAccessor) ParsePath(path string) (
	*accessors.OSPath, error) {

	fmt.Println("[VMFS ACCESSOR] ParsePath returns NOT nil")
	return accessors.NewESXiVMFSPath(path)
}

func (self VMFSFileSystemAccessor) ReadDir(filename string) (
	[]accessors.FileInfo, error) {

	fmt.Println("[VMFS ACCESSOR] ReadDir returns nil")
	return nil, nil
}

func (self *VMFSFileSystemAccessor) ReadDirWithOSPath(
	fullpath *accessors.OSPath) (res []accessors.FileInfo, err error) {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Printf("PANIC %v\n", r)
			debug.PrintStack()
			err, _ = r.(error)
		}
	}()

	// result := []accessors.FileInfo{}

	vmfs_ctx, err := GetVMFSContext(self.scope, self.device, fullpath, self.accessor)
	if err != nil {
		return nil, err
	}

	fmt.Println("VMFS ACCESSOR] ReadDirWithOSPath vmfs_ctx", vmfs_ctx.descriptor.Magic)

	fmt.Println("[VMFS ACCESSOR] ReadDirWithOSPath returns nil")
	return nil, nil
}

func (self *readAdapter) ReadAt(buf []byte, offset int64) (int, error) {
	self.Lock()
	defer self.Unlock()
	self.pos = offset

	fmt.Println("[VMFS ACCESSOR] ReadAt returns NOT nil")
	return self.reader.ReadAt(buf, offset)
}

func (self *readAdapter) Ranges() []uploads.Range {
	result := []uploads.Range{}
	for _, rng := range self.reader.Ranges() {
		result = append(result, uploads.Range{
			Offset: rng.Offset,
			Length: rng.Length,
		})
	}
	return result
}

func init() {
	accessors.Register("raw_vmfs", &VMFSFileSystemAccessor{},
		`Access a VMFS filesystem.`)
}
