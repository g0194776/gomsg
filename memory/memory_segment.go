package memory

import "fmt"

//MemorySegment used for taking fixed size of allocated memory from OS.
//It supports lots of atomic operations e.g. Write or Read.
type memorySegment struct {
	data          []byte
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
	if uint(len(value)) > ms.bytesLeft {
		panic(fmt.Sprintf("BUG: specified data size larger than left size. (Bytes Needed: %d, Bytes Left: %d)", len(value), ms.bytesLeft))
	}
	copy(ms.data[ms.usedOffset:], value)
	ms.usedOffset += uint(len(value))
	ms.bytesLeft -= uint(len(value))
	return nil
}
