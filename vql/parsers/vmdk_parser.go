package parsers

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/Velocidex/ordereddict"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"www.velocidex.com/golang/velociraptor/accessors"
	"www.velocidex.com/golang/velociraptor/acls"
	"www.velocidex.com/golang/velociraptor/vql"
	vql_subsystem "www.velocidex.com/golang/velociraptor/vql"
	vfilter "www.velocidex.com/golang/vfilter"
	"www.velocidex.com/golang/vfilter/arg_parser"
)

const (
	SectorSize        = 512
	UTF16LECharacters = 72
	SignaturePosition = 8
	PartitionGUID     = 16
	DiskGUIDSize      = 16
)

var file *accessors.OSPath

// LBA 1
type GPTHeader struct {
	Signature                [SignaturePosition]byte
	Revision                 uint32
	HeaderSize               uint32
	HeaderCRC32              uint32
	Reserved                 uint32
	CurrentLBA               uint64
	BackupLBA                uint64
	FirstUsableLBA           uint64
	LastUsableLBA            uint64
	DiskGUID                 [DiskGUIDSize]byte
	StartingLBAEntries       uint64
	NumberOfPartitionEntries uint32
	SizeOfPartitionEntry     uint32
	PartitionEntriesCRC32    uint32
}

// LBA 2-33
type GPTPartitionEntry struct {
	PartitionTypeGUID   [PartitionGUID]byte
	UniquePartitionGUID [PartitionGUID]byte
	FirstLBA            uint64
	LastLBA             uint64
	Attributes          uint64
	PartitionName       [UTF16LECharacters]byte
}

type VmdkParserArgs struct {
	Filenames []*accessors.OSPath `vfilter:"required,field=filename,doc=A list of log files to parse."`
	Accessor  string              `vfilter:"optional,field=accessor,doc=The accessor to use."`
}

type VmdkParser struct{}

func (self VmdkParser) Info(scope vfilter.Scope, type_map *vfilter.TypeMap) *vfilter.PluginInfo {
	return &vfilter.PluginInfo{
		Name:     "vmdk_parser",
		Doc:      "parses Sparse VMDK files.",
		ArgType:  type_map.AddType(scope, &VmdkParser{}),
		Metadata: vql.VQLMetadata().Permissions(acls.PREPARE_RESULTS).Build(),
		Version:  1,
	}
}

func (self VmdkParser) Call(ctx context.Context,
	scope vfilter.Scope,
	args *ordereddict.Dict) <-chan vfilter.Row {
	output_chan := make(chan vfilter.Row)

	fmt.Println("[VMDK_PARSER] VMDK Parser called")

	go func() {
		defer close(output_chan)
		defer vql_subsystem.RegisterMonitor("vmdk_parser", args)()

		arg := &VmdkParserArgs{}
		err := arg_parser.ExtractArgsWithContext(ctx, scope, args, arg)
		if err != nil {
			fmt.Println("[VMDK_PARSER] Error extracting args: ", err.Error())
			scope.Log("[CONCAT]: %s", err.Error())
			return
		}

		fmt.Println("[VMDK_PARSER] Path: ", arg.Filenames)

		err = vql_subsystem.CheckFilesystemAccess(scope, arg.Accessor)
		if err != nil {
			fmt.Println("[VMDK_PARSER] Error checking filesystem access: ", err.Error())
			return
		}

		fmt.Println("[VMDK_PARSER] Filesystem access checked")
		fmt.Println("[VMDK_PARSER] Current System: ", os.Getenv("OS"))

		accessor, err := accessors.GetAccessor(arg.Accessor, scope)
		if err != nil {
			fmt.Println("[VMDK_PARSER] Error getting accessor: ", err.Error())
			return
		}

		for _, filename := range arg.Filenames {
			file = filename
			fd, err := accessor.OpenWithOSPath(file)
			if err != nil {
				fmt.Println("[VMDK_PARSER] Error opening file: ", err.Error())
				return
			}
			defer fd.Close()

			fmt.Println("[VMDK_PARSER] File opened: ", filename.String())

			buff := bytes.NewBuffer([]byte{})

			fmt.Println("[VMDK_PARSER] Buffer created")

			size, err := io.Copy(buff, fd)

			fmt.Println("[VMDK_PARSER] File size: ", size)

			if err != nil {
				fmt.Println("[VMDK_PARSER] Error reading file: ", err.Error())
				return
			}

			fmt.Println("[VMDK_PARSER] File read into buffer")

			reader := bytes.NewReader(buff.Bytes())
			fmt.Println("[VMDK_PARSER] File read: ", size)

			header, err := parseGPTHeader(reader)
			if err != nil {
				fmt.Println("[VMDK_PARSER] Error parsing GPT header: ", err.Error())
				return
			}

			fmt.Println("[VMDK_PARSER] GPT header parsed")

			partitions, err := parseGPTPartitionEntries(reader, header)
			if err != nil {
				fmt.Println("[VMDK_PARSER] Error parsing GPT partition entries: ", err.Error())
				return
			}

			fmt.Println("[VMDK_PARSER] GPT partition entries parsed")
			fmt.Println("[VMDK_PARSER] Number of partitions: ", len(partitions))

			for _, entry := range partitions {
				fmt.Println("Partition Entry: ", decodeUTF16(entry.PartitionName[:]))
				select {
				case <-ctx.Done():
					return
				case output_chan <- partitionEntryToMap(entry):
				}
			}
		}
	}()

	return output_chan

}

func init() {
	vql_subsystem.RegisterPlugin(&VmdkParser{})
}

func partitionEntryToMap(entry GPTPartitionEntry) map[string]interface{} {
	return map[string]interface{}{
		"PartitionTypeGUID":   fmt.Sprintf("%x", entry.PartitionTypeGUID),
		"UniquePartitionGUID": fmt.Sprintf("%x", entry.UniquePartitionGUID),
		"FirstLBA":            entry.FirstLBA,
		"LastLBA":             entry.LastLBA,
		"Attributes":          entry.Attributes,
		"PartitionName":       decodeUTF16(entry.PartitionName[:]),
	}
}

func decodeUTF16(data []byte) string {
	decoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
	utf8, _ := io.ReadAll(transform.NewReader(bytes.NewReader(data), decoder))
	return string(utf8)
}

func parseGPTHeader(r io.ReaderAt) (*GPTHeader, error) {
	buf := make([]byte, SectorSize)

	if _, err := r.ReadAt(buf, SectorSize); err != nil {
		return nil, fmt.Errorf("failed to read GPT header: %v", err)
	}

	fmt.Println("GPT Header: ", buf)
	fmt.Printf("%c\n", buf[:8])

	if string(buf[:8]) != "EFI PART" {
		return nil, fmt.Errorf("GPT header signature not found")
	}

	// All are little endian: https://en.wikipedia.org/wiki/GUID_Partition_Table
	header := &GPTHeader{
		Revision:                 binary.LittleEndian.Uint32(buf[8:12]),
		HeaderSize:               binary.LittleEndian.Uint32(buf[12:16]),
		HeaderCRC32:              binary.LittleEndian.Uint32(buf[16:20]),
		Reserved:                 binary.LittleEndian.Uint32(buf[20:24]),
		CurrentLBA:               binary.LittleEndian.Uint64(buf[24:32]),
		BackupLBA:                binary.LittleEndian.Uint64(buf[32:40]),
		FirstUsableLBA:           binary.LittleEndian.Uint64(buf[40:48]),
		LastUsableLBA:            binary.LittleEndian.Uint64(buf[48:56]),
		StartingLBAEntries:       binary.LittleEndian.Uint64(buf[72:80]),
		NumberOfPartitionEntries: binary.LittleEndian.Uint32(buf[80:84]),
		SizeOfPartitionEntry:     binary.LittleEndian.Uint32(buf[84:88]),
		PartitionEntriesCRC32:    binary.LittleEndian.Uint32(buf[88:92]),
	}

	copy(header.Signature[:], buf[:8])
	copy(header.DiskGUID[:], buf[56:72])

	return header, nil
}

func parseGPTPartitionEntries(r io.ReaderAt, header *GPTHeader) ([]GPTPartitionEntry, error) {
	num_entries := header.NumberOfPartitionEntries
	size_entry := header.SizeOfPartitionEntry
	total_size := int64(num_entries) * int64(size_entry)
	entries_data := make([]byte, total_size)
	start_offset := int64(header.StartingLBAEntries) * SectorSize
	if _, err := r.ReadAt(entries_data, start_offset); err != nil {
		return nil, fmt.Errorf("failed to read partition entries: %v", err)
	}

	var partitions []GPTPartitionEntry
	for i := 0; i < int(num_entries); i++ {

		// for every 128 bytes, we have a partition entry
		offset := i * int(size_entry)

		// check if the entry is empty
		entry_data := entries_data[offset : offset+int(size_entry)]
		is_empty := true
		for j := 0; j < 16; j++ {
			if entry_data[j] != 0 {
				is_empty = false
				break
			}
		}

		// if entry isn't empty, it's a legit partition that can be parsed
		if !is_empty {
			var entry GPTPartitionEntry
			copy(entry.PartitionTypeGUID[:], entry_data[0:16])
			copy(entry.UniquePartitionGUID[:], entry_data[16:32])
			entry.FirstLBA = binary.LittleEndian.Uint64(entry_data[32:40])
			entry.LastLBA = binary.LittleEndian.Uint64(entry_data[40:48])
			entry.Attributes = binary.LittleEndian.Uint64(entry_data[48:56])
			copy(entry.PartitionName[:], entry_data[56:128])
			partitions = append(partitions, entry)
		}
	}
	return partitions, nil
}
