package memory

import (
	"errors"
	"sync"

	log "github.com/sirupsen/logrus"
	"sync/atomic"
)

var (
	//total memeory pool size for allocating from OS by default.
	//default size: 100Mi
	defMemPoolSize uint = 1024 * 1024 * 100
	//size of each memory segment for taking the real memory.
	defmemSegmentSize uint = 256
)

//MemoryProvider providers lots of abilities for managing memory usages internal.
type MemoryProvider struct {
	memPool []byte
	//usedSegments []*memorySegment
	unusedSegmentHead  *memorySegment
	unusedSegmentCount *int32
	sync.RWMutex
}

//Initialize memory pool.
//Passing ZERO(0) will use default values to initializes memory pool.
func (mp *MemoryProvider) Initialize(memPoolSize, memSegmentSize uint) {
	mps := uint(0)
	mss := uint(0)
	if memPoolSize == 0 {
		mps = defMemPoolSize
	} else {
		mps = memPoolSize
	}
	if memSegmentSize == 0 {
		mss = defmemSegmentSize
	} else {
		mss = memSegmentSize
	}
	log.Infof("Initializing Memory Pool, Size: %d", mps)
	multiples := mps / mss
	mp.memPool = make([]byte, 0, mps)
	var initCnt int32 = 0
	mp.unusedSegmentCount = &initCnt
	//mp.usedSegments = make([]*memorySegment, 0, multiples)
	for index := 0; index < int(multiples); index++ {
		//segment raw data.
		data := mp.memPool[index*int(mss) : (index*int(mss))+int(mss)]
		ms := &memorySegment{
			data:          data,
			rawDataOffset: uint(index) * mss,
			usedOffset:    0,
			SegmentLength: mss,
			bytesLeft:     mss,
			CurrentStatus: MEM_SEGMENT_STATUS_INIT}
		mp.Giveback(ms)
	}
}

func (mp *MemoryProvider) NewSegmentProxy() MemorySegmentProxyer {
	return &MemorySegmentProxy{mp: mp,
		usedSegments: []*memorySegment{}}
}

//GetOneAvailable method returns an in-used memory segment.
//If there isn't any avaiable memory segment, it'll returns an error immediatelly.
func (mp *MemoryProvider) GetOneAvailable() (*memorySegment, error) {
	mp.Lock()
	defer mp.Unlock()
	if *mp.unusedSegmentCount == 0 {
		return nil, errors.New("No more available memory segments can be use.")
	}
	ms := mp.unusedSegmentHead
	ms.CurrentStatus = MEM_SEGMENT_STATUS_BORROWED
	mp.unusedSegmentHead = ms.Previous
	//decrease counter.
	atomic.AddInt32(mp.unusedSegmentCount, -1)
	return ms, nil
}

//Giveback an in-used memory segment.
func (mp *MemoryProvider) Giveback(ms *memorySegment) error {
	if ms == nil {
		return errors.New("Nil Pointer being passed.")
	}
	if ms.CurrentStatus != MEM_SEGMENT_STATUS_BORROWED && ms.CurrentStatus != MEM_SEGMENT_STATUS_INIT {
		return errors.New("CANNOT give the same memory segment more than once!")
	}
	mp.Lock()
	defer mp.Unlock()
	ms.CurrentStatus = MEM_SEGMENT_STATUS_POOLING
	ms.usedOffset = 0
	ms.bytesLeft = ms.SegmentLength
	if *mp.unusedSegmentCount == 0 {
		mp.unusedSegmentHead = ms
	} else {
		ms.Previous = mp.unusedSegmentHead
		mp.unusedSegmentHead = ms
	}
	//increase counter.
	atomic.AddInt32(mp.unusedSegmentCount, 1)
	return nil
}
