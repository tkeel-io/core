package util

const (
	// 比 65536 小的最大素数.
	mod = 65521
	// nmax is the largest n such that.
	// 255 * n * (n+1) / 2 + (n+1) * (mod-1) <= 2^32-1.
	// It is mentioned in RFC 1950 (search for "5552").
	nmax = 5552
)

type digest uint32

func (d *digest) Reset() { *d = 1 }

func (d *digest) Write(p []byte) (nn int, err error) {
	*d = update(*d, p)
	return len(p), nil
}

// Add p to the running checksum d.
func update(d digest, p []byte) digest {
	s1, s2 := uint32(d&0xffff), uint32(d>>16)
	for len(p) > 0 {
		var q []byte
		if len(p) > nmax {
			p, q = p[:nmax], p[nmax:]
		}

		for len(p) >= 4 {
			s1 += uint32(p[0])
			s2 += s1
			s1 += uint32(p[1])
			s2 += s1
			s1 += uint32(p[2])
			s2 += s1
			s1 += uint32(p[3])
			s2 += s1
			p = p[4:]
		}
		for _, x := range p {
			s1 += uint32(x)
			s2 += s1
		}
		s1 %= mod
		s2 %= mod
		p = q
	}

	return digest(s2<<16 | s1)
}

func Hash32(text string) uint32 {
	var instance = digest(1)
	instance.Write([]byte(text))
	return uint32(instance)
}
