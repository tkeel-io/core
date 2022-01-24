package util

import "testing"

func TestHash32(t *testing.T) {
	var d digest
	d.Write([]byte("test for adler"))
	t.Log(d)
}

func BenchmarkHash32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Hash32("too yong too simple.")
	}
}
