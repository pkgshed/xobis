package xobis

import "strconv"

func appendCode(dst []byte, c Groups, present Presence) []byte {
	if present.PresentA() {
		dst = strconv.AppendUint(dst, uint64(c.A), 10)
		dst = append(dst, '-')
	}
	if present.PresentB() {
		dst = strconv.AppendUint(dst, uint64(c.B), 10)
		dst = append(dst, ':')
	}
	if present.PresentC() {
		dst = strconv.AppendUint(dst, uint64(c.C), 10)
	}
	if present.PresentD() {
		dst = append(dst, '.')
		dst = strconv.AppendUint(dst, uint64(c.D), 10)
	}
	if present.PresentE() {
		dst = append(dst, '.')
		dst = strconv.AppendUint(dst, uint64(c.E), 10)
	}
	if present.PresentF() {
		dst = append(dst, '*')
		dst = strconv.AppendUint(dst, uint64(c.F), 10)
	}
	return dst
}
