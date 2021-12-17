/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package constraint

type BitBucket struct {
	lenth  int
	bucket []uint8
}

func NewBitBucket(length int) *BitBucket {
	byteLen := (length + 7) / 8
	return &BitBucket{
		lenth:  length,
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
	if bb.lenth <= n {
		panic("index overflow")
	}
	return n / 8, n % 8
}
