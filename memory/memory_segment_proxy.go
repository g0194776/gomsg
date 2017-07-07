package memory

import (
	"bytes"
	"encoding/binary"
	"math"
)

type MemorySegmentProxyer interface {
	WriteInt32(value int32) error
	WriteUInt32(value uint32) error
	WriteInt64(value int64) error
	WriteUInt64(value uint64) error
	GetBuffer() []byte
	WriteString(value string) error
	WriteMemory(data []byte) error
	//GetBuffer() ([]byte, error)
}

type MemorySegmentProxy struct {
	mp           *MemoryProvider
	usedSegments []*memorySegment
}

func (msp *MemorySegmentProxy) WriteInt32(value int32) error {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(value))
	return msp.WriteMemory(data)
}

func (msp *MemorySegmentProxy) WriteUInt32(value uint32) error {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, value)
	return msp.WriteMemory(data)
}

func (msp *MemorySegmentProxy) WriteInt64(value int64) error {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(value))
	return msp.WriteMemory(data)
}

func (msp *MemorySegmentProxy) WriteUInt64(value uint64) error {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, value)
	return msp.WriteMemory(data)
}

func (msp *MemorySegmentProxy) WriteString(value string) error {
	if value == "" {
		return nil
	}
	data := []byte(value)
	return msp.WriteMemory(data)
}

func (msp *MemorySegmentProxy) WriteMemory(data []byte) error {
	bytesLeft := len(data)
	mss, nestedSegmentUsing, err := msp.getAvailableSegment(uint(bytesLeft))
	if err != nil {
		return err
	}
	if len(mss) == 1 {
		mss[0].WriteBytes(data)
		if !nestedSegmentUsing {
			msp.usedSegments = append(msp.usedSegments, mss[0])
		}
	} else {
		currentOffset := 0
		for i := 0; i < len(mss); i++ {
			bytesWritten := msp.calcBytesCount(bytesLeft, int(mss[i].bytesLeft))
			err := mss[i].WriteBytes(data[currentOffset:bytesWritten])
			currentOffset += bytesWritten
			bytesLeft -= bytesWritten
			if err != nil {
				return err
			}
			//Avoid using the same segment twice during calling the GetBuffer() method.
			if nestedSegmentUsing && i == 0 {
				continue
			}
			msp.usedSegments = append(msp.usedSegments, mss[i])
		}
	}
	return nil
}

func (msp *MemorySegmentProxy) calcBytesCount(bytesLeft, segmentBytesLeft int) int {
	bytesWritten := 0
	//Calc how much data SHOULD be write into the memory segment.
	if bytesLeft >= int(defmemSegmentSize) {
		if segmentBytesLeft >= int(defmemSegmentSize) {
			bytesWritten = int(defmemSegmentSize)
		} else {
			bytesWritten = segmentBytesLeft
		}
	} else {
		if segmentBytesLeft >= bytesLeft {
			bytesWritten = bytesLeft
		} else {
			bytesWritten = segmentBytesLeft
		}
	}
	return bytesWritten
}

func (msp *MemorySegmentProxy) getAvailableSegment(size uint) ([]*memorySegment, bool, error) {
	var lastUsedSegment *memorySegment = nil
	if len(msp.usedSegments) != 0 {
		lastUsedSegment = msp.usedSegments[len(msp.usedSegments)-1]
	}
	if lastUsedSegment != nil && lastUsedSegment.HasEnoughMemory(size) {
		return msp.usedSegments[len(msp.usedSegments)-1:], true, nil
	}
	nestedSegmentUsing := false
	//estimate how many memory segments will be use
	var totalSize int
	if lastUsedSegment != nil {
		totalSize = int(lastUsedSegment.bytesLeft) + int(size)
	} else {
		totalSize = int(size)
	}

	segmentCnt := int(math.Ceil(float64(totalSize) / float64(defmemSegmentSize)))
	segments := make([]*memorySegment, 0, segmentCnt)
	if lastUsedSegment != nil && lastUsedSegment.bytesLeft != 0 {
		nestedSegmentUsing = true
		segments = append(segments, lastUsedSegment)
		segmentCnt--
	}
	//Memory segments allocation.
	for i := 0; i < segmentCnt; i++ {
		seg, err := msp.mp.GetOneAvailable()
		if err != nil {
			return nil, false, err
		}
		segments = append(segments, seg)
	}
	return segments, nestedSegmentUsing, nil
}

func (msp *MemorySegmentProxy) GetBuffer() []byte {
	if msp.usedSegments == nil || len(msp.usedSegments) == 0 {
		return []byte{}
	}
	buff := &bytes.Buffer{}
	for _, seg := range msp.usedSegments {
		if seg.bytesLeft == 0 {
			//Writes full memory data.
			buff.Write(seg.data)
		} else {
			//Writes used memory data.
			buff.Write(seg.data[:seg.usedOffset])
		}
		//free used memory segment.
		msp.mp.Giveback(seg)
	}
	return buff.Bytes()
}
