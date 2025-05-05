package vmfs

import (
	"encoding/binary"
	"fmt"
	"io"
)

type FS_CONFIG uint32

const (
	FS_CONFIG_MAINTENANCE FS_CONFIG = 0x00000008
	FS_CONFIG_SYSTEM      FS_CONFIG = 0x00000800
)

type FDS_VolInfo struct {
	ID [32]byte
}

type FS3_Checksum struct {
	Value       uint64
	ChecksumGen uint64
}

type FS3FileDescriptor struct {
	Magic                  uint32
	MajorVersion           uint32
	MinorVersion           uint8
	UUID                   [16]byte
	Config                 FS_CONFIG
	FSLabel                [128]byte
	DiskBlockSize          uint32
	FileBlockSize          uint64
	CreationTime           uint32
	SnapID                 uint32
	VolInfo                FDS_VolInfo
	FDCClusterGroupOffset  uint32
	FDCClustersPerGroup    uint32
	SubBlockSize           uint32
	MaxJournalSlotsPerTxn  uint32
	PB2VolAddr             uint64
	PB2FDAddr              uint32
	HostUUID               [16]byte
	GBLGeneration          uint64
	SDDVolAddr             uint64
	SDDFDAddr              uint32
	ChecksumType           uint8
	UnmapPriority          uint16
	Pad1                   [4]byte
	ChecksumGen            uint64
	Checksum               FS3_Checksum
	PhysDiskBlockSize      uint32
	MDAlignment            uint32
	SFBToLFBShift          uint16
	Reserved16_1           uint16
	Reserved16_2           uint16
	PtrBlockShift          uint16
	SFBAddrBits            uint16
	Reserved16_3           uint16
	TBZGranularity         uint32
	JournalBlockSize       uint32
	LeaseIntervalMs        uint32
	ReclaimWindowMs        uint32
	LocalStampUS           uint64
	LocalMountOwnerMacAddr [6]byte
}

func ReadFS3Descriptor(r io.ReaderAt) (*FS3FileDescriptor, error) {
	buf := make([]byte, FS3DescriptorSize)
	_, err := r.ReadAt(buf, FS3Offset)
	if err != nil {
		return nil, err
	}

	desc := &FS3FileDescriptor{}
	err = binary.Read(io.NewSectionReader(r, FS3Offset, FS3DescriptorSize), binary.LittleEndian, desc)
	if err != nil {
		return nil, err
	}

	if desc.Magic != FS3Magic && desc.Magic != FS3Magic+1 {
		return nil, fmt.Errorf("invalid FS3 magic: 0x%x", desc.Magic)
	}

	return desc, nil
}
