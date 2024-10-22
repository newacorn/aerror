package aerror

import (
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

const (
	errSplit                 = "=>"
	errSplitNewLine          = errSplit + "\n"
	stacktraceLen            = 6
	stackWrapperStart        = "[["
	stackWrapperEnd          = "]]"
	fileLineIndent           = 4
	filePathStripCount       = 3
	stackWrapperEndMultiline = stackWrapperEnd + "\n"
)

var errSplitWithNewLineBytes = []byte(errSplitNewLine)

var spaces = [100]byte{
	' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
	' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
	' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
	' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
	' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
	' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
	' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
	' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
	' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
	' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
}

type Error struct {
	msg string
	pcs frames
	err error
}

func New(msg string) error {
	ps := make([]uintptr, 8)
	return &Error{msg: msg, pcs: ps[:runtime.Callers(2, ps)]}
}

func With(err error, msg string) error {
	ps := make([]uintptr, stacktraceLen)
	return &Error{err: err, msg: msg, pcs: ps[:runtime.Callers(2, ps)]}
}

func (e *Error) Error() string {
	if e.err != nil {
		return b2s(append(append(e.pcs.Format(e.msg, 0, false), errSplit...), e.err.Error()...))
	}
	return b2s(e.pcs.Format(e.msg, 0, false))
}

func (e *Error) MultiLine(indents ...int) string {
	var indent int
	if len(indents) > 0 {
		indent = indents[0]
	}
	b := e.pcs.Format(e.msg, indent, true)
	if e.err == nil {
		return b2s(b)
	}
	er, ok := e.err.(MultiLiner)
	if ok {
		return b2s(append(append(b, splitWithPrefixSpace(indent)...), er.MultiLine(indent+len(errSplit))...))
	}
	return b2s(append(append(b, splitWithPrefixSpace(indent)...), append(space(len(errSplit)), e.err.Error()...)...))
}

type frames []uintptr

func (fs frames) Format(msg string, indent int, multiline bool) []byte {
	bufSize := 512
	if multiline {
		bufSize = 768
	}
	buf := make([]byte, 0, bufSize)
	stackIndent := indent + len(stackWrapperEnd)
	if multiline {
		buf = append(buf, space(indent)...)
	}
	buf = append(buf, stackWrapperStart[0])
	//msg = color.RenderString(color.New(color.Red, color.Bold).String(), msg)
	buf = append(buf, msg...)
	buf = append(buf, stackWrapperStart[1])
	if multiline {
		buf = append(buf, '\n')
	}
	cf := runtime.CallersFrames(fs)
	start := false
	for {
		f, more := cf.Next()
		if !more {
			break
		}
		if start {
			if multiline {
				buf = append(buf, '\n')
			} else {
				buf = append(buf, '|')
			}
		}
		if multiline {
			buf = append(buf, space(stackIndent)...)
		}
		buf = append(buf, f.Function...)
		buf = append(buf, ':')
		if multiline {
			buf = append(buf, '\n')
			buf = append(buf, space(stackIndent+fileLineIndent)...)
			buf = append(buf, f.File...)
		} else {
			buf = append(buf, pathSuffix(f.File, filePathStripCount)...)
		}
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(f.Line), 10)
		start = true
	}
	if multiline {
		buf = append(buf, '\n')
		buf = append(buf, space(indent)...)
		buf = append(buf, stackWrapperEndMultiline...)
		return buf
	}
	return append(buf, stackWrapperEnd...)
}

func b2s(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func s2b(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func pathSuffix(path string, slashCount int) (suffix string) {
	var idx = -1
	for i := 0; i <= slashCount; i++ {
		idx2 := strings.Index(path, "/")
		if idx2 == -1 {
			if idx == -1 {
				return path
			}
			break
		}
		idx = idx2 + 1
		path = path[idx:]
	}
	return path
}

func space(count int) []byte {
	if count <= 100 {
		return spaces[:count]
	}
	b1 := make([]byte, count)
	for i := 0; i < count; i++ {
		b1[i] = ' '
	}
	return b1
}

func splitWithPrefixSpace(count int) []byte {
	if count == 0 {
		return errSplitWithNewLineBytes
	}
	buf := make([]byte, 0, len(errSplitNewLine)+count)
	buf = append(buf, space(count)...)
	buf = append(buf, errSplit...)
	buf = append(buf, '\n')
	return buf
}

type MultiLiner interface {
	MultiLine(indents ...int) string
}
