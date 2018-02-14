package memory

import (
	"bytes"
	"fmt"
	"math"
)

type MemorySegmentProxyer interface {
	WriteInt32(value int32, serialization_func func(v int32) ([]byte, error)) error
	WriteUInt32(value uint32, serialization_func func(v uint32) ([]byte, error)) error
	WriteInt64(value int64, serialization_func func(v int64) ([]byte, error)) error
	WriteUInt64(value uint64, serialization_func func(v uint64) ([]byte, error)) error
	WriteString(value string, serialization_func func(v string) ([]byte, error)) error
	GetBuffer() []byte
	GetPosition() *MemoryPosition
	Skip(cnt uint) error
	GetSegmentCount() int
	Close()
}

var (
	ErrSerializationFuncMissed = fmt.Errorf("serialization function is required.")
)

type MemorySegmentProxy struct {
	mp           *MemoryProvider
	usedSegments []*memorySegment
}

func (msp *MemorySegmentProxy) GetSegmentCount() int {
	return len(msp.usedSegments)
}

func (msp *MemorySegmentProxy) WriteInt32(value int32, serialization_func func(v int32) ([]byte, error)) error {
	mss, err := msp.getAvailableSegment(INT32_SIZE)
	if err != nil {
		return err
	}
	if len(mss) == 1 {
		mss[0].WriteInt32(value)
		return nil
	}
	if serialization_func == nil {
		return ErrSerializationFuncMissed
	}
	data, err := serialization_func(value)
	if err != nil {
		return err
	}
	return msp.WriteMemoryToSegments(data, mss)
}

func (msp *MemorySegmentProxy) WriteUInt32(value uint32, serialization_func func(v uint32) ([]byte, error)) error {
	mss, err := msp.getAvailableSegment(INT32_SIZE)
	if err != nil {
		return err
	}
	if len(mss) == 1 {
		mss[0].WriteUInt32(value)
		return nil
	}
	if serialization_func == nil {
		return ErrSerializationFuncMissed
	}
	data, err := serialization_func(value)
	if err != nil {
		return err
	}
	return msp.WriteMemoryToSegments(data, mss)
}

func (msp *MemorySegmentProxy) WriteInt64(value int64, serialization_func func(v int64) ([]byte, error)) error {
	mss, err := msp.getAvailableSegment(INT64_SIZE)
	if err != nil {
		return err
	}
	if len(mss) == 1 {
		mss[0].WriteInt64(value)
		return nil
	}
	if serialization_func == nil {
		return ErrSerializationFuncMissed
	}
	data, err := serialization_func(value)
	if err != nil {
		return err
	}
	return msp.WriteMemoryToSegments(data, mss)
}

func (msp *MemorySegmentProxy) WriteUInt64(value uint64, serialization_func func(v uint64) ([]byte, error)) error {
	mss, err := msp.getAvailableSegment(INT64_SIZE)
	if err != nil {
		return err
	}
	if len(mss) == 1 {
		mss[0].WriteUInt64(value)
		return nil
	}
	if serialization_func == nil {
		return ErrSerializationFuncMissed
	}
	data, err := serialization_func(value)
	if err != nil {
		return err
	}
	return msp.WriteMemoryToSegments(data, mss)
}

func (msp *MemorySegmentProxy) WriteString(value string, serialization_func func(v string) ([]byte, error)) error {
	if value == "" {
		return nil
	}
	data, err := serialization_func(value)
	if err != nil {
		return err
	}
	return msp.WriteMemory(data)
}

func (msp *MemorySegmentProxy) WriteMemory(data []byte) error {
	bytesLeft := len(data)
	mss, err := msp.getAvailableSegment(uint(bytesLeft))
	if err != nil {
		return err
	}
	if len(mss) == 1 {
		mss[0].WriteBytes(data)
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
		}
	}
	return nil
}

func (msp *MemorySegmentProxy) WriteMemoryToSegments(data []byte, mss []*memorySegment) error {
	bytesLeft := len(data)
	if len(mss) == 1 {
		mss[0].WriteBytes(data)
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

func (msp *MemorySegmentProxy) getAvailableSegment(size uint) ([]*memorySegment, error) {
	bytesLeft := 0
	if len(msp.usedSegments) != 0 {
		if msp.usedSegments[len(msp.usedSegments)-1].HasEnoughMemory(size) {
			return msp.usedSegments[len(msp.usedSegments)-1:], nil
		} else {
			bytesLeft = int(msp.usedSegments[len(msp.usedSegments)-1].bytesLeft)
		}
	}

	//
	//SITUATION 1, nothing left on the last of segment.
	//
	//	segment-size  = 8
	//	required-size = 4
	//--------------------------------------------------
	//
	//            seg1
	//|xxxxxxxxxxxxxxxxxxxxxxxxxxxx| <-- fully used
	//            seg2
	//|----------------------------| <-- new segment starts here.

	startSegmentIndex := 0
	if bytesLeft == 0 {
		//next element.
		startSegmentIndex = len(msp.usedSegments)
	} else {
		startSegmentIndex = len(msp.usedSegments) - 1
	}
	//estimate how many memory segments will be use
	totalSize := bytesLeft + int(size)
	segmentCnt := int(math.Ceil(float64(totalSize) / float64(defmemSegmentSize)))
	//Memory segments allocation.
	for i := 0; i < segmentCnt; i++ {
		seg, err := msp.mp.GetOneAvailable()
		if err != nil {
			return nil, err
		}
		msp.usedSegments = append(msp.usedSegments, seg)
	}
	return msp.usedSegments[startSegmentIndex:], nil
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

func (msp *MemorySegmentProxy) Skip(cnt uint) error {
	mss, err := msp.getAvailableSegment(cnt)
	if err != nil {
		return err
	}
	if len(mss) == 1 {
		mss[0].Skip(cnt)
		return nil
	}
	bytesLeft := cnt
	for i := 0; i < len(mss); i++ {
		if mss[i].bytesLeft <= bytesLeft {
			mss[i].Skip(mss[i].bytesLeft)
			bytesLeft -= mss[i].bytesLeft
		} else {
			mss[i].Skip(bytesLeft)
		}
	}
	return nil
}

func (msp *MemorySegmentProxy) GetPosition() *MemoryPosition {
	mp := &MemoryPosition{}
	if len(msp.usedSegments) == 0 {
		mp.SegmentIndex = 0
		mp.SegmentOffset = 0
	} else {
		mp.SegmentIndex = len(msp.usedSegments) - 1
		mp.SegmentOffset = int(msp.usedSegments[len(msp.usedSegments)-1].usedOffset)
	}
	return mp
}

func (msp *MemorySegmentProxy) Close() {
	if len(msp.usedSegments) > 0 {
		for _, s := range msp.usedSegments {
			msp.mp.Giveback(s)
		}
		//clear set.
		msp.usedSegments = nil
	}
}
