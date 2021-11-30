package constraint

const (
	EnabledFlagSelf = 1 << iota
	EnabledFlagSearch
	EnabledFlagTimeSeries
)

type BitBucket struct {
	length int
	bucket []uint8
}

func NewBitBucket(length int) *BitBucket {
	byteLen := (length + 7) / 8
	return &BitBucket{
		length: length,
		bucket: make([]uint8, byteLen),
	}
}

func (bb *BitBucket) Enabled(n int) bool {
	byteIndex, remIndex := bb.indexable(n)
	val, offsetFlag := bb.bucket[byteIndex], uint8(1<<remIndex)
	return val&offsetFlag > 0
}

func (bb *BitBucket) Enable(n int) bool {
	byteIndex, remIndex := bb.indexable(n)
	oldValue, offsetFlag := bb.bucket[byteIndex], uint8(1<<remIndex)
	bb.bucket[byteIndex] |= offsetFlag
	return oldValue&offsetFlag > 0
}

func (bb *BitBucket) Disable(n int) bool {
	byteIndex, remIndex := bb.indexable(n)
	oldValue, offsetFlag := bb.bucket[byteIndex], uint8(1<<remIndex)
	bb.bucket[byteIndex] &= ^offsetFlag
	return oldValue&offsetFlag > 0
}

func (bb *BitBucket) indexable(n int) (int, int) {
	if bb.length <= n {
		panic("index overflow")
	}
	return n / 8, n % 8
}
