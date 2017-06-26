package memory

import "bytes"

//MemorySegment used for taking fixed size of allocated memory from OS.
//It supports lots of atomic operations e.g. Write or Read.
type memorySegment struct {
	data          []byte
	dataBuff      *bytes.Buffer
	usedOffset    uint
	rawDataOffset uint
	SegmentLength uint
	CurrentStatus uint
	bytesLeft     uint
}

const (
	MEM_SEGMENT_STATUS_POOLING = iota
	MEM_SEGMENT_STATUS_BORROWED
)

func (ms *memorySegment) HasEnoughMemory(memorySize uint) bool {
	return (ms.SegmentLength - ms.usedOffset) >= memorySize
}

func (ms *memorySegment) WriteBytes(value []byte) error {
	return nil
}
