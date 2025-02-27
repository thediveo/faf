// Copyright 2024 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package faf

import "bytes"

// Bytestring provides efficient parsing of text lines in form of byte slices,
// assuming contents to be in ASCII and treating UTF-8 as individual byte-sized
// characters.
//
// We rely on the compiler emitting efficient machine code for determining
// len(b), so we don't need to store it explicitly, but use the length already
// stored as part of the byte slice anyway. Looking at the generated x86-64
// machine code using a tools like the “[Compiler Explorer]” confirms efficient
// code where len(bs.b) results in a single MOVQ instruction. For arm64 “all we
// get” is also a single MOVD instruction.
//
// [Compiler Explorer]: https://godbolt.org/
type Bytestring struct {
	b   []byte // line contents
	pos int    // parsing position within the line contents
}

// NewBytestring returns a new Bytestring object for parsing the supplied text
// line as a byte slice. It is small enough to be allocated on the stack,
// avoiding heap allocations.
func NewBytestring(b []byte) *Bytestring {
	return &Bytestring{
		pos: 0,
		b:   b,
	}
}

// EOL returns true if the parsing has reached the end of the byte string,
// otherwise false.
func (b *Bytestring) EOL() (eol bool) { return b.pos >= len(b.b) }

// SkipSpace skips over any space 0x20 characters until either reaching the
// first non-space character, or EOF. When reaching EOL, it returns true.
func (b *Bytestring) SkipSpace() (eol bool) {
	for {
		if b.pos >= len(b.b) {
			return true
		}
		if b.b[b.pos] != ' ' {
			return false
		}
		b.pos++
	}
}

// SkipText skips the text s in the buffer at the current position if present,
// returning ok true. Otherwise, returns ok false and the buffer's parsing
// position is left unchanged.
func (b *Bytestring) SkipText(s string) (ok bool) {
	if b.pos >= len(b.b) || b.pos+len(s) > len(b.b) {
		return false
	}
	if !bytes.Equal([]byte(s), b.b[b.pos:b.pos+len(s)]) {
		return false
	}
	b.pos += len(s)
	return true
}

// Next returns the next byte from the bytestring and ok true, otherwise ok
// false.
func (b *Bytestring) Next() (ch byte, ok bool) {
	if b.pos >= len(b.b) {
		return 0, false
	}
	ch = b.b[b.pos]
	b.pos++
	return ch, true
}

const cutoffDecimalUint64 = (1<<64-1)/10 + 1

// Uint64 parses the decimal number starting in the buffer at the current
// position until a character other than 0-9 is encountered, or EOL. The number
// must consist of at least a single digit. If successful, Uint64 returns the
// number and true; otherwise zero and false. Overflowing Uint64 is also
// considered to be an error, returning zero and false in this case.
func (b *Bytestring) Uint64() (num uint64, ok bool) {
	for {
		if b.pos >= len(b.b) {
			if !ok {
				// We never consumed at least a single digit, so this is right
				// dead on arrival.
				return 0, false
			}
			// Reached the end and we had at least a single digit consumed, so
			// this is fine.
			return num, true
		}
		ch := b.b[b.pos]
		if ch < '0' || ch > '9' {
			if !ok {
				// Again, the first character is already bad, so we report an
				// error.
				return 0, false
			}
			// We've reached the end of the number, other stuff now following;
			// we're done and successfully report the number we've parsed.
			return num, true
		}
		// Don't overflow...
		if num >= cutoffDecimalUint64 {
			return 0, false
		}
		num = num*10 + uint64(ch-'0')
		b.pos++
		ok = true // yes, we successfully got a(nother) digit.
	}
}

const cutoffHexUint64 = 1 << 60

// HexUint64 parses the hexadecimal number starting in the buffer at the current
// position until a character other than 0-9, a-f, or A-F is encountered, or
// EOL. The number must consist of at least a single hex digit. If successful,
// HexUint64 returns the number and true; otherwise zero and false. Overflowing
// HexUint64 is also considered to be an error, returning zero and false in this
// case.
func (b *Bytestring) HexUint64() (num uint64, ok bool) {
	for {
		if b.pos >= len(b.b) {
			if !ok {
				// We never consumed at least a single digit, so this is right
				// dead on arrival.
				return 0, false
			}
			// Reached the end and we had at least a single digit consumed, so
			// this is fine.
			return num, true
		}
		ch := b.b[b.pos]
		var digit byte
		if ch >= '0' && ch <= '9' {
			digit = ch - '0'
		} else if ch >= 'a' && ch <= 'f' {
			digit = ch - 'a' + 10
		} else if ch >= 'A' && ch <= 'F' {
			digit = ch - 'A' + 10
		} else if !ok {
			return 0, false
		} else {
			// We've reached the end of the number, other stuff now following;
			// we're done and successfully report the number we've parsed.
			return num, true
		}
		// Don't overflow...
		if num >= cutoffHexUint64 {
			return 0, false
		}
		num = num<<4 + uint64(digit)
		b.pos++
		ok = true // yes, we successfully got a(nother) digit.
	}
}

// NumFields returns the number of fields found in the line, starting from the
// current position. NumFields does not change the current position. Fields are
// made of sequences of characters excluding the space character. Fields are
// separated by one or more spaces.
func (b *Bytestring) NumFields() (num int) {
	pos := b.pos
	for {
		for {
			if pos >= len(b.b) {
				return
			}
			if b.b[pos] != ' ' {
				break
			}
			pos++
		}
		num++
		for {
			if pos >= len(b.b) {
				return
			}
			if b.b[pos] == ' ' {
				break
			}
			pos++
		}
	}
}
