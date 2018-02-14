package memory

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

//MemorySegment used for taking fixed size of allocated memory from OS.
//It supports lots of atomic operations e.g. Write or Read.
type memorySegment struct {
	data          []byte
	usedOffset    uint
	rawDataOffset uint
	SegmentLength uint
	CurrentStatus uint
	bytesLeft     uint
	Previous      *memorySegment
}

type MemorySegmentWriter interface {
	WriteInt32(value int32) error
	WriteUInt32(value uint32) error
	WriteInt64(value int64) error
	WriteUInt64(value uint64) error
	GetBuffer() []byte
	WriteString(value string) error
	WriteMemory(data []byte) error
	Skip(length uint) error
	//GetBuffer() ([]byte, error)
}

const (
	MEM_SEGMENT_STATUS_INIT = iota
	MEM_SEGMENT_STATUS_POOLING
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

func (ms *memorySegment) WriteInt32(value int32) error {
	binary.LittleEndian.PutUint32(ms.data[ms.usedOffset:], uint32(value))
	ms.usedOffset += INT32_SIZE
	ms.bytesLeft -= INT32_SIZE
	return nil
}

func (ms *memorySegment) WriteUInt32(value uint32) error {
	binary.LittleEndian.PutUint32(ms.data[ms.usedOffset:], value)
	ms.usedOffset += INT32_SIZE
	ms.bytesLeft -= INT32_SIZE
	return nil
}

func (ms *memorySegment) WriteInt64(value int64) error {
	binary.LittleEndian.PutUint64(ms.data[ms.usedOffset:], uint64(value))
	ms.usedOffset += INT64_SIZE
	ms.bytesLeft -= INT64_SIZE
	return nil
}

func (ms *memorySegment) WriteUInt64(value uint64) error {
	binary.LittleEndian.PutUint64(ms.data[ms.usedOffset:], uint64(value))
	ms.usedOffset += INT64_SIZE
	ms.bytesLeft -= INT64_SIZE
	return nil
}

func (ms *memorySegment) GetBuffer() []byte {
	panic("Please DO NOT directly call this method from a MemorySegment object.")
}

func (ms *memorySegment) WriteString(value string) error {
	buf := bytes.NewBuffer(ms.data[ms.usedOffset:])
	buf.WriteString(value)
	ms.usedOffset += uint(len(value))
	ms.bytesLeft -= uint(len(value))
	return nil
}

func (ms *memorySegment) WriteMemory(data []byte) error {
	copy(ms.data[ms.usedOffset:], data)
	ms.usedOffset += uint(len(data))
	ms.bytesLeft -= uint(len(data))
	return nil
}

func (ms *memorySegment) Skip(length uint) error {
	ms.usedOffset += length
	ms.bytesLeft -= length
	return nil
}
