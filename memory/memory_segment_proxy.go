package memory

import "encoding/binary"

type MemorySegmentProxyer interface {
	WriteInt32(value int32) error
	//GetBuffer() ([]byte, error)
}

type MemorySegmentProxy struct {
	mp              *MemoryProvider
	usedSegments    []*memorySegment
	lastUsedSegment *memorySegment
}

func (msp *MemorySegmentProxy) WriteInt32(value int32) error {
	mss, err := msp.getAvailableSegment(4)
	if err != nil {
		return err
	}
	data := make([]byte, 0, 4)
	binary.LittleEndian.PutUint32(data, uint32(value))
	if len(mss) == 1 {
		mss[0].WriteBytes(data)
	} else {
		currentOffset := 0
		bytesLeft := 4
		for i := 0; i < len(mss); i++ {
			bytesWritten := 0
			segmentBytesLeft := int(mss[i].bytesLeft)
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

func (msp *MemorySegmentProxy) getAvailableSegment(size int) ([]*memorySegment, error) {
	return nil, nil
}
