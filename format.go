package main

import (
	"encoding/base64"
	"encoding/hex"

	prng "github.com/EricLagergren/go-prng/xorshift"
)

func encode(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func isSpecial(b byte) bool {
	return specialTable.in(b)
}

func isDigit(b byte) bool {
	return '0' <= b && b <= '9'
}

func isUpper(b byte) bool {
	return 'A' <= b && b <= 'Z'
}

func isLower(b byte) bool {
	return 'a' <= b && b <= 'z'
}

// format takes in a buffer of (hopefully) sufficiently random
// data and returns a printable string conforming to the command-
// line constraints.
func format(buf []byte, b64 bool) string {
	if len(buf) < *length {
		exit("format received a buffer with length %d, wanted %d", len(buf), *length)
	}

	tmp := make([]byte, hex.EncodedLen(len(buf)))
	hex.Encode(tmp, buf)

	dst := make([]byte, *length)
	start := int(r.Next() % uint64(len(tmp)))

	n := copy(dst, tmp[start:])

	have := len(tmp) - start

	// Wrap around again
	if have < *length {
		copy(dst[n:], tmp[:*length-have])
	}

	dig, low := 0, 0
	for i := range dst {
		if isDigit(dst[i]) {
			dig++
		}

		if isLower(dst[i]) {
			low++
		}
	}

	if *lower > 0 {
		lowerTable.add(dst, *lower, low, isLower)
	}
	if *special > 0 {
		specialTable.add(dst, *special, 0, isSpecial)
	}
	if *upper > 0 {
		upperTable.add(dst, *upper, 0, isUpper)
	}
	if *digits > 0 && dig < *digits {
		digitTable.add(dst, *digits, dig, isDigit)
	}

	if b64 {
		return encode(dst)
	}
	return string(dst)
}

var (
	r       = &prng.Shift4096Star{}
	visited []bool
	initial = true
)

func init() { r.Seed() }

func next(x int) uint64 { return r.Next() % uint64(x) }

func (t table) add(buf []byte, need, have int, want func(byte) bool) {

	if initial {
		visited = make([]bool, *length)
		initial = false
	}

	for i := next(len(buf)); have < need; i = next(len(buf)) {
		if !visited[i] && !want(buf[i]) {
			buf[i] = t.get()
			visited[i] = true
			have++
		}
	}
}
