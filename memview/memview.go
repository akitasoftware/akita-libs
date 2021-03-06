package memview

import (
	"bytes"
	"io"
)

// MemView represents a "view" on a collection of byte slices. Conceptually, you
// may think of it as a [][]byte, with helper methods to make it seem like one
// contiguous []byte. It is designed to help minimize the amount of copying when
// dealing with large buffers of data.
//
// Modifying a MemView does not change the underlying data. Instead, it simply
// changes the pointers to where to read data from.
//
// Copying a MemView or passing memView by value is like copying a slice - it's
// efficient, but modifications to the copy affect the original MemView and vice
// versa. Use `DeepCopy` to create a completely independent MemView.
//
// The zero value is an empty MemView ready to use.
type MemView struct {
	buf    [][]byte
	length int64
}

// The new MemView does NOT make a copy of data, so the caller MUST ensure that
// the underlying memory of data remains valid and unmodified after this call
// returns.
func New(data []byte) MemView {
	return MemView{
		buf:    [][]byte{data},
		length: int64(len(data)),
	}
}

func (dst *MemView) Append(src MemView) {
	dst.buf = append(dst.buf, src.buf...)
	dst.length += src.length
}

// Creates a MemView that is completely independent from the current one.
func (mv MemView) DeepCopy() MemView {
	newBuf := make([][]byte, len(mv.buf))
	copy(newBuf, mv.buf)
	return MemView{
		buf:    newBuf,
		length: mv.length,
	}
}

func (mv *MemView) CreateReader() *MemViewReader {
	return &MemViewReader{mv: mv}
}

func (mv *MemView) Clear() {
	mv.buf = mv.buf[:0] // clear without reallocating memory
	mv.length = 0
}

func (mv MemView) Len() int64 {
	return mv.length
}

// Returns the byte at the given index. Returns 0 if index is out of bounds.
func (mv MemView) GetByte(index int64) byte {
	if index < 0 {
		return 0
	}

	n := index
	for i := 0; i < len(mv.buf); i++ {
		lb := int64(len(mv.buf[i]))
		if n < lb {
			return mv.buf[i][n]
		}
		n -= lb
	}
	return 0
}

// Returns mv[start:end] (end is not inclusive). Returns an empty MemView if
// range is invalid.
func (mv MemView) SubView(start, end int64) MemView {
	if start >= end {
		return MemView{}
	}

	startBuf := -1
	endBuf := -1
	var startOffset, endOffset int

	var n int64
	for i, b := range mv.buf {
		lb := int64(len(b))
		if startBuf == -1 && n+lb > start {
			startBuf = i
			startOffset = int(start - n)
		}
		if endBuf == -1 && n+lb >= end { // >= because end is not inclusive
			endBuf = i
			endOffset = int(end - n)
			break
		}
		n += lb
	}

	if startBuf == -1 || endBuf == -1 {
		return MemView{}
	}

	newBuf := make([][]byte, endBuf+1-startBuf)
	copy(newBuf, mv.buf[startBuf:endBuf+1])
	newMS := MemView{
		buf:    newBuf,
		length: end - start,
	}
	if len(newMS.buf) == 1 {
		newMS.buf[0] = newMS.buf[0][startOffset:endOffset]
	} else {
		newMS.buf[0] = newMS.buf[0][startOffset:]
		newMS.buf[len(newMS.buf)-1] = newMS.buf[len(newMS.buf)-1][:endOffset]
	}
	return newMS
}

// Index returns the index of the first instance of sep in mv after start index,
// or -1 if sep is not present in mv.
func (mv MemView) Index(start int64, sep []byte) int64 {
	// Find the first buffer to start from.
	startBuf := -1
	startOffset := 0
	var currIndex int64
	for i, b := range mv.buf {
		lb := int64(len(b))
		if currIndex+lb-1 < start { // -1 because start is an index
			currIndex += lb
		} else {
			startBuf = i
			startOffset = int(start - currIndex)
			currIndex += int64(startOffset)
			break
		}
	}

	if startBuf == -1 {
		return -1
	} else if len(sep) == 0 {
		return start
	}

	// Iteratively search for the target, keeping in mind that the target may be
	// spread over multiple slices in mv.buf.
	needle := sep
	var foundIndex int64
	for b := startBuf; b < len(mv.buf) && len(needle) > 0; b++ {
		haystack := mv.buf[b][startOffset:]

		searchLen := len(needle)
		if len(haystack) < searchLen {
			searchLen = len(haystack)
		}

		match := true
		for i := 0; i < searchLen; i++ {
			if haystack[i] != needle[i] {
				// Reset the search.
				match = false
				needle = sep
				if i < len(haystack)-1 {
					// We have more hay in the current hay stack to check.
					b--
					startOffset += i + 1
					currIndex += int64(i + 1)
				} else {
					// Move on to the next buf in mv.buf.
					currIndex += int64(len(haystack))
					startOffset = 0
				}
				break
			}
		}

		if match {
			if len(needle) == len(sep) {
				// This is the initial find.
				foundIndex = currIndex
			}
			needle = needle[searchLen:]

			// Move on to the next buf in mv.buf.
			currIndex += int64(len(haystack))
			startOffset = 0
		}
	}

	if len(needle) == 0 {
		return foundIndex
	}

	return -1
}

// Returns a string of all the data referenced by this MemView. Note that is
// creates a COPY of the underlying data.
func (mv MemView) String() string {
	var buf bytes.Buffer
	io.Copy(&buf, mv.CreateReader())
	return string(buf.Bytes())
}

type MemViewReader struct {
	mv *MemView

	// Index for the element from mv.buf to read next.
	rIndex int

	// Offest into mv.buf[rIndex] for the next read.
	rOffset int
}

// If MemView has no data to return, err is io.EOF (unless len(out) is zero),
// otherwise it is nil. This behavior matches that of bytes.Buffer.
func (r *MemViewReader) Read(out []byte) (int, error) {
	if len(out) == 0 {
		return 0, nil
	} else if r.rIndex >= len(r.mv.buf) { // really just ==, but use >= to be safer
		return 0, io.EOF
	}

	bytesRead := 0
	for i := r.rIndex; i < len(r.mv.buf); i++ {
		curr := r.mv.buf[i][r.rOffset:]
		cp := copy(out[bytesRead:], curr)
		bytesRead += cp
		if cp == len(curr) {
			r.rIndex += 1
			r.rOffset = 0
		} else {
			// If cp < len(curr), it means we've run out of output space.
			r.rOffset += cp
			return bytesRead, nil
		}
	}

	// We've read something, so don't return EOF in case more data gets passed to
	// this MemView.
	return bytesRead, nil
}

// Make MemView more efficient as a source in io.Copy.
func (r *MemViewReader) WriteTo(dst io.Writer) (int64, error) {
	var bytesWritten int64
	for _, b := range r.mv.buf {
		n, err := dst.Write(b)
		bytesWritten += int64(n)
		if err != nil {
			return bytesWritten, err
		}
	}
	return bytesWritten, nil
}
