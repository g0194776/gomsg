package memory

import (
	"testing"

	assert "github.com/lexandro/go-assert"
)

func TestInitializeMemoryPool(t *testing.T) {
	mp := &MemoryProvider{}
	mp.Initialize(0, 0) //use default value
	assert.IsTrue(t, cap(mp.memPool) == int(defMemPoolSize))
	assert.IsTrue(t, len(mp.unusedSegments) == int(defMemPoolSize)/int(defmemSegmentSize))
}

func TestInitializeMemoryPool_WithCustomizedParameters(t *testing.T) {
	mp := &MemoryProvider{}
	mp.Initialize(1024, 128) //use default value
	assert.IsTrue(t, cap(mp.memPool) == 1024)
	assert.IsTrue(t, len(mp.unusedSegments) == 1024/128)
}

func TestInitializeMemoryPool_WithCustomizedParameters_GetAvailableSegmentSucceed(t *testing.T) {
	mp := &MemoryProvider{}
	mp.Initialize(128, 64) //use default value
	assert.IsTrue(t, cap(mp.memPool) == 128)
	assert.IsTrue(t, len(mp.unusedSegments) == 2)

	ms, err := mp.GetOneAvailable()
	assert.IsNotNil(t, ms)
	assert.IsNil(t, err)
	assert.IsTrue(t, len(mp.unusedSegments) == 1)
	//assert.IsTrue(t, len(mp.usedSegments) == 1)
}

func TestInitializeMemoryPool_WithCustomizedParameters_GetAllAvailableSegmentsSucceed(t *testing.T) {
	mp := &MemoryProvider{}
	mp.Initialize(128, 64) //use default value
	assert.IsTrue(t, cap(mp.memPool) == 128)
	assert.IsTrue(t, len(mp.unusedSegments) == 2)

	ms, err := mp.GetOneAvailable()
	assert.IsNotNil(t, ms)
	assert.IsNil(t, err)
	assert.IsTrue(t, len(mp.unusedSegments) == 1)
	//assert.IsTrue(t, len(mp.usedSegments) == 1)

	//Execute Gets method again.
	ms, err = mp.GetOneAvailable()
	assert.IsNotNil(t, ms)
	assert.IsNil(t, err)

	//No more available memory segments.
	ms, err = mp.GetOneAvailable()
	assert.IsTrue(t, ms == nil)
	assert.IsNotNil(t, err)
}

func TestInitializeMemoryPool_WithCustomizedParameters_GivebackTwice(t *testing.T) {
	mp := &MemoryProvider{}
	mp.Initialize(128, 64) //use default value
	assert.IsTrue(t, cap(mp.memPool) == 128)
	assert.IsTrue(t, len(mp.unusedSegments) == 2)

	ms, err := mp.GetOneAvailable()
	assert.IsNotNil(t, ms)
	assert.IsNil(t, err)
	assert.IsTrue(t, len(mp.unusedSegments) == 1)
	//assert.IsTrue(t, len(mp.usedSegments) == 1)

	//Execute Gets method again.
	ms2, err := mp.GetOneAvailable()
	assert.IsNotNil(t, ms2)
	assert.IsNil(t, err)

	//No more available memory segments.
	ms3, err := mp.GetOneAvailable()
	assert.IsTrue(t, ms3 == nil)
	assert.IsNotNil(t, err)
	assert.IsTrue(t, len(mp.unusedSegments) == 0)

	assert.IsNil(t, mp.Giveback(ms))
	assert.IsTrue(t, len(mp.unusedSegments) == 1)
	//Giveback the same memory segment more than once.
	assert.IsNotNil(t, mp.Giveback(ms))
	assert.IsNil(t, mp.Giveback(ms2))
	assert.IsTrue(t, len(mp.unusedSegments) == 2)
}

func TestInitializeMemoryPool_WithCustomizedParameters_GivebackNilPointer(t *testing.T) {
	mp := &MemoryProvider{}
	mp.Initialize(128, 64) //use default value
	assert.IsTrue(t, cap(mp.memPool) == 128)
	assert.IsTrue(t, len(mp.unusedSegments) == 2)

	ms, err := mp.GetOneAvailable()
	assert.IsNotNil(t, ms)
	assert.IsNil(t, err)
	assert.IsTrue(t, len(mp.unusedSegments) == 1)
	//assert.IsTrue(t, len(mp.usedSegments) == 1)

	assert.IsNotNil(t, mp.Giveback(nil))
}
