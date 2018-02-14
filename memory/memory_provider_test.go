package memory

import (
	"testing"

	. "gopkg.in/check.v1"
)

type MemoryPool struct{}

var _ = Suite(&MemoryPool{})

func (m *MemoryPool) TestInitializeMemoryPool(c *C) {
	mp := &MemoryProvider{}
	mp.Initialize(0, 0) //use default value
	c.Assert(cap(mp.memPool), Equals, int(defMemPoolSize))
	c.Assert(*mp.unusedSegmentCount, Equals, int32(int32(defMemPoolSize)/int32(defmemSegmentSize)))
}

func (m *MemoryPool) TestInitializeMemoryPool_WithCustomizedParameters(c *C) {
	mp := &MemoryProvider{}
	mp.Initialize(1024, 128) //use default value
	c.Assert(cap(mp.memPool), Equals, 1024)
	c.Assert(*mp.unusedSegmentCount, Equals, int32(1024/128))
}

func (m *MemoryPool) TestInitializeMemoryPool_WithCustomizedParameters_GetAvailableSegmentSucceed(c *C) {
	mp := &MemoryProvider{}
	mp.Initialize(128, 64) //use default value
	c.Assert(cap(mp.memPool), Equals, 128)
	c.Assert(*mp.unusedSegmentCount, Equals, int32(2))

	ms, err := mp.GetOneAvailable()
	c.Assert(ms, NotNil)
	c.Assert(err, IsNil)
	c.Assert(*mp.unusedSegmentCount, Equals, int32(1))
	//assert.IsTrue(t, len(mp.usedSegments) == 1)
}

func (m *MemoryPool) TestInitializeMemoryPool_WithCustomizedParameters_GetAllAvailableSegmentsSucceed(c *C) {
	mp := &MemoryProvider{}
	mp.Initialize(128, 64) //use default value
	c.Assert(cap(mp.memPool), Equals, 128)
	c.Assert(*mp.unusedSegmentCount, Equals, int32(2))

	ms, err := mp.GetOneAvailable()
	c.Assert(ms, NotNil)
	c.Assert(err, Equals, nil)
	c.Assert(*mp.unusedSegmentCount, Equals, int32(1))
	//assert.IsTrue(t, len(mp.usedSegments) == 1)

	//Execute Gets method again.
	ms, err = mp.GetOneAvailable()
	c.Assert(ms, NotNil)
	c.Assert(err, Equals, nil)

	//No more available memory segments.
	ms, err = mp.GetOneAvailable()
	c.Assert(ms, IsNil)
	c.Assert(err, NotNil)
}

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

func (m *MemoryPool) TestInitializeMemoryPool_WithCustomizedParameters_GivebackTwice(c *C) {
	mp := &MemoryProvider{}
	mp.Initialize(128, 64) //use default value
	c.Assert(cap(mp.memPool), Equals, 128)
	c.Assert(*mp.unusedSegmentCount, Equals, int32(2))

	ms, err := mp.GetOneAvailable()
	c.Assert(ms, NotNil)
	c.Assert(err, IsNil)
	c.Assert(*mp.unusedSegmentCount, Equals, int32(1))
	//assert.IsTrue(t, len(mp.usedSegments) == 1)

	//Execute Gets method again.
	ms2, err := mp.GetOneAvailable()
	c.Assert(ms2, NotNil)
	c.Assert(err, IsNil)

	//No more available memory segments.
	ms3, err := mp.GetOneAvailable()
	c.Assert(ms3, IsNil)
	c.Assert(err, NotNil)
	c.Assert(*mp.unusedSegmentCount, Equals, int32(0))

	c.Assert(mp.Giveback(ms), IsNil)
	c.Assert(*mp.unusedSegmentCount, Equals, int32(1))
	//Giveback the same memory segment more than once.
	c.Assert(mp.Giveback(ms), NotNil)
	c.Assert(mp.Giveback(ms2), IsNil)
	c.Assert(*mp.unusedSegmentCount, Equals, int32(2))
}

func (m *MemoryPool) TestInitializeMemoryPool_WithCustomizedParameters_GivebackNilPointer(c *C) {
	mp := &MemoryProvider{}
	mp.Initialize(128, 64) //use default value
	c.Assert(cap(mp.memPool), Equals, 128)
	c.Assert(*mp.unusedSegmentCount, Equals, int32(2))

	ms, err := mp.GetOneAvailable()
	c.Assert(ms, NotNil)
	c.Assert(err, IsNil)
	c.Assert(*mp.unusedSegmentCount, Equals, int32(1))
	//assert.IsTrue(t, len(mp.usedSegments) == 1)

	c.Assert(mp.Giveback(nil), NotNil)
}
