package memory

import (
	"fmt"

	"github.com/gomsg/serializations"
	. "gopkg.in/check.v1"
)

type MemoryProxy struct{}

var _ = Suite(&MemoryProxy{})

func (m *MemoryProxy) Test_Raw_WriteInt32_Succeed(c *C) {
	mp := &MemoryProvider{}
	mp.Initialize(24, 8)
	c.Assert(cap(mp.memPool), Equals, 24)
	c.Assert(*mp.unusedSegmentCount, Equals, int32(3))
	segmentProxy := mp.NewSegmentProxy()
	c.Assert(segmentProxy, NotNil)
	err := segmentProxy.WriteInt32(10, serializations.INT32_SERIALIZATION)
	if err != nil {
		c.Fatal(err)
	}
	c.Assert(*mp.unusedSegmentCount, Equals, int32(2))
	err = segmentProxy.WriteInt32(11, serializations.INT32_SERIALIZATION)
	if err != nil {
		c.Fatal(err)
	}
	c.Assert(*mp.unusedSegmentCount, Equals, int32(2))
	data := segmentProxy.GetBuffer()
	c.Assert(data, NotNil)
	fmt.Printf("Segment Length: %d\n", len(data))
	fmt.Printf("Segment Data: [%# x]\n", data)
	c.Assert(len(data), Equals, 8)
	c.Assert(*mp.unusedSegmentCount, Equals, int32(3))
}

func (m *MemoryProxy) Test_Skip(c *C) {
	mp := &MemoryProvider{}
	mp.Initialize(128, 32)
	c.Assert(cap(mp.memPool), Equals, 128)
	c.Assert(*mp.unusedSegmentCount, Equals, int32(4))
	//
	//SITUATION: A new memory pool has been initialized.
	//
	//	segment-size  = 128
	//	required-size = 32
	//--------------------------------------------------
	//
	//            seg1
	//|----------------------------|
	//            seg2
	//|----------------------------|
	//            seg3
	//|----------------------------|
	//            seg4
	//|----------------------------|
	msp := mp.NewSegmentProxy().(*MemorySegmentProxy)
	c.Assert(msp, NotNil)
	c.Assert(len(msp.usedSegments), Equals, 0)
	//
	//SITUATION: A new memory proxy has been allocated.
	//
	//	segment-size  = 128
	//	required-size = 32
	//--------------------------------------------------
	//There is no any segment to be used for storing data.

	msp.Skip(4)
	//
	//	segment-size  = 128
	//	required-size = 32
	//
	//	*  x - write back bytes.
	//	*  y - string bytes
	//	*  □ - un-use bytes.
	//--------------------------------------------------
	//
	//            seg1
	//|□□□□------------------------|
	c.Assert(len(msp.usedSegments), Equals, 1)
	c.Assert(msp.usedSegments[0].usedOffset, Equals, uint(4))
	msp.Skip(4)
	//
	//	segment-size  = 128
	//	required-size = 32
	//
	//	*  x - write back bytes.
	//	*  y - string bytes
	//	*  □ - un-use bytes.
	//--------------------------------------------------
	//
	//            seg1
	//|□□□□□□□□--------------------|
	c.Assert(len(msp.usedSegments), Equals, 1)
	c.Assert(msp.usedSegments[0].usedOffset, Equals, uint(8))
	msp.Skip(24)
	//
	//	segment-size  = 128
	//	required-size = 32
	//
	//	*  x - write back bytes.
	//	*  y - string bytes
	//	*  □ - un-use bytes.
	//--------------------------------------------------
	//
	//            seg1
	//|□□□□□□□□□□□□□□□□□□□□□□□□□□□□| <--fully used.
	c.Assert(len(msp.usedSegments), Equals, 1)
	c.Assert(msp.usedSegments[0].usedOffset, Equals, uint(32))
	msp.Skip(0) // <- doesn't effect anything.
	//
	//	segment-size  = 128
	//	required-size = 32
	//
	//	*  x - write back bytes.
	//	*  y - string bytes
	//	*  □ - un-use bytes.
	//--------------------------------------------------
	//
	//            seg1
	//|□□□□□□□□□□□□□□□□□□□□□□□□□□□□| <--fully used.

	msp.Skip(1)
	//
	//	segment-size  = 128
	//	required-size = 32
	//
	//	*  x - write back bytes.
	//	*  y - string bytes
	//	*  □ - un-use bytes.
	//--------------------------------------------------
	//
	//            seg1
	//|□□□□□□□□□□□□□□□□□□□□□□□□□□□□| <--fully used.
	//            seg2
	//|□---------------------------|
	c.Assert(len(msp.usedSegments), Equals, 2)
	c.Assert(msp.usedSegments[0].usedOffset, Equals, uint(32))
	c.Assert(msp.usedSegments[1].usedOffset, Equals, uint(1))
	c.Assert(*mp.unusedSegmentCount, Equals, int32(2))

	//close.
	msp.Close()
	c.Assert(len(msp.usedSegments), Equals, 0)
	c.Assert(*mp.unusedSegmentCount, Equals, int32(4))
}

func (m *MemoryProxy) Test_GetPosition(c *C) {
	mp := &MemoryProvider{}
	mp.Initialize(128, 32)
	c.Assert(cap(mp.memPool), Equals, 128)
	c.Assert(*mp.unusedSegmentCount, Equals, int32(4))
	msp := mp.NewSegmentProxy().(*MemorySegmentProxy)
	c.Assert(msp, NotNil)
	c.Assert(len(msp.usedSegments), Equals, 0)
	pos := msp.GetPosition()
	c.Assert(pos, NotNil)
	c.Assert(pos.SegmentIndex, Equals, 0)
	c.Assert(pos.SegmentOffset, Equals, 0)
	c.Assert(len(msp.usedSegments), Equals, 0)

	msp.Skip(4)
	//
	//	segment-size  = 128
	//	required-size = 32
	//
	//	*  x - write back bytes.
	//	*  y - string bytes
	//	*  □ - un-use bytes.
	//--------------------------------------------------
	//
	//            seg1
	//|□□□□------------------------|

	pos = msp.GetPosition()
	c.Assert(pos, NotNil)
	c.Assert(pos.SegmentIndex, Equals, 0)
	c.Assert(pos.SegmentOffset, Equals, 4)

	msp.Skip(28)
	//
	//	segment-size  = 128
	//	required-size = 32
	//
	//	*  x - write back bytes.
	//	*  y - string bytes
	//	*  □ - un-use bytes.
	//--------------------------------------------------
	//
	//            seg1
	//|□□□□□□□□□□□□□□□□□□□□□□□□□□□□| <--fully used.
	pos = msp.GetPosition()
	c.Assert(pos, NotNil)
	c.Assert(pos.SegmentIndex, Equals, 0)
	c.Assert(pos.SegmentOffset, Equals, 32)

	msp.Skip(4)
	//
	//	segment-size  = 128
	//	required-size = 32
	//
	//	*  x - write back bytes.
	//	*  y - string bytes
	//	*  □ - un-use bytes.
	//--------------------------------------------------
	//
	//            seg1
	//|□□□□□□□□□□□□□□□□□□□□□□□□□□□□| <--fully used.
	//            seg2
	//|□□□□------------------------|
	c.Assert(len(msp.usedSegments), Equals, 2)
	pos = msp.GetPosition()
	c.Assert(pos, NotNil)
	c.Assert(pos.SegmentIndex, Equals, 1)
	c.Assert(pos.SegmentOffset, Equals, 4)
	//clear resource.
	msp.Close()
}
