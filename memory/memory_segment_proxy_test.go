package memory

import (
	"fmt"
	"testing"

	assert "github.com/lexandro/go-assert"
)

func Test_Raw_WriteInt32_Succeed(t *testing.T) {
	mp := &MemoryProvider{}
	mp.Initialize(24, 8)
	assert.IsTrue(t, cap(mp.memPool) == 24)
	assert.IsTrue(t, len(mp.unusedSegments) == 3)
	segmentProxy := mp.NewSegmentProxy()
	assert.IsNotNil(t, segmentProxy)
	err := segmentProxy.WriteInt32(10)
	if err != nil {
		t.Fatal(err)
	}
	assert.IsTrue(t, len(mp.unusedSegments) == 2)
	err = segmentProxy.WriteInt32(11)
	if err != nil {
		t.Fatal(err)
	}
	assert.IsTrue(t, len(mp.unusedSegments) == 2)
	data := segmentProxy.GetBuffer()
	assert.IsNotNil(t, data)
	fmt.Printf("Segment Length: %d\n", len(data))
	fmt.Printf("Segment Data: [%# x]\n", data)
	assert.IsTrue(t, len(data) == 8)
	assert.IsTrue(t, len(mp.unusedSegments) == 3)
}
