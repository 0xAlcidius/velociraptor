package vmfs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"www.velocidex.com/golang/velociraptor/accessors"
)

type VMFSContext struct {
	reader     io.ReaderAt
	descriptor *FS3Descriptor
	rootPath   *accessors.OSPath
}

func NewRawVMFS(reader io.ReaderAt) (*VMFSContext, error) {
	section := io.NewSectionReader(reader, DescriptorOffset, 512)

	fmt.Println("[VMFS ACCESSOR] NewRawVMFS - Calling ReadDescriptor")
	descriptor, err := ReadDescriptor(section, 0)
	if err != nil {
		return nil, fmt.Errorf("Failed to read VMFS FS3 descriptor: %w", err)
	}

	return &VMFSContext{
		reader:     reader,
		descriptor: descriptor,
	}, nil
}

type FS3Descriptor struct {
	Magic               uint32
	MajorVersion        uint32
	MinorVersion        uint8
	UUID                [16]byte
	Config              uint32
	FSLabel             [128]byte
	DiskBlockSize       uint32
	FileBlockSize       uint64
	CreationTime        uint32
	SnapID              uint32
	VolInfo             [32]byte
	FDCClusterOffset    uint32
	FDCClustersPerGroup uint32
	SubBlockSize        uint32
	MaxJournalSlots     uint32
	PB2VolAddr          uint64
	PB2FDAddr           uint32
	HostUUID            [16]byte
	GlobalGen           uint64
	SDDVolAddr          uint64
	SDDFDAddr           uint32
	ChecksumType        uint8
	UnmapPriority       uint16
	Padding             [4]byte
	ChecksumGen         uint64
	Checksum            [16]byte
}

type FS3FileDescriptor struct {
	Address          uint32
	Generation       uint32
	NumLinks         uint32
	Type             uint32
	Flags            uint32
	Length           uint64
	BlockSize        uint64
	NumBlocks        uint64
	ModTime          uint32
	CreationTime     uint32
	AccessTime       uint32
	UID              uint32
	GID              uint32
	Mode             uint32
	ZLA              uint32
	TBZLo            uint32
	COWLo            uint32
	NewEpochLo       uint32
	TBZHi            uint32
	COWHi            uint32
	NumPointerBlocks uint32
	NewEpochHi       uint32
	_unk1            uint32
	AffinityFD       uint32
	TBZGranularity   uint32
	ParentFD         uint32
}

type FS3DirEntry struct {
	Type       uint32
	Address    uint32
	Generation uint32
	Name       [128]byte
}

func ReadDescriptor(rd io.ReaderAt, offset int64) (*FS3Descriptor, error) {
	fmt.Println("[VMFS ACCESSOR] ReadDescriptor")
	buf := make([]byte, DescriptorSize)
	rd.ReadAt(buf, offset)
	var desc *FS3Descriptor

	fmt.Println("[VMFS ACCESSOR] ReadDescriptor Read first bytes, magic: ", buf[0:4])

	err := binary.Read(bytes.NewReader(buf), binary.LittleEndian, &desc)
	if err != nil {
		return nil, err
	}

	fmt.Println("[VMFS ACCESSOR] ReadDescriptor Read descriptor, magic: ", desc.Magic)

	if desc.Magic != DescriptorMagic && desc.Magic != DescriptorMagic+1 {
		return nil, fmt.Errorf("invalid FS3 magic: 0x%x", desc.Magic)
	}

	return desc, nil
}

func ReadFileDescriptor(r io.ReaderAt, offset int64) (*FS3FileDescriptor, error) {
	const diskLockSize = 512
	var fd FS3FileDescriptor

	size := binary.Size(fd)
	if size < 0 {
		return nil, fmt.Errorf("invalid file descriptor size")
	}

	buf := make([]byte, size)
	_, err := r.ReadAt(buf, offset+diskLockSize)
	if err != nil {
		return nil, err
	}

	err = binary.Read(bytes.NewReader(buf), binary.LittleEndian, &fd)
	if err != nil {
		return nil, err
	}

	return &fd, nil
}

func ListDir(r io.ReaderAt, fd *FS3FileDescriptor, baseOffset int64) error {
	numEntries := int(fd.Length) / DirEntrySize

	buf := make([]byte, fd.Length)
	_, err := r.ReadAt(buf, baseOffset)
	if err != nil {
		return err
	}

	for i := 0; i < numEntries; i++ {
		entryBuf := buf[i*DirEntrySize : (i+1)*DirEntrySize]
		var dir FS3DirEntry
		b := bytes.NewReader(entryBuf)
		if err := binary.Read(b, binary.LittleEndian, &dir); err != nil {
			continue
		}

		name := string(bytes.TrimRight(dir.Name[:], "\x00"))
		if dir.Address != 0 {
			fmt.Printf("Entry: %s (Type: %d, Addr: 0x%x)\n", name, dir.Type, dir.Address)
		}
	}

	return nil
}

type Range struct {
	Offset, Length int64
}

type RangeReaderAt interface {
	io.ReaderAt

	Ranges() []Range
}
