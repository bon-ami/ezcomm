package ezcomm

import (
	"bytes"
	"encoding/binary"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitee.com/bon-ami/eztools/v4"
)

/* bytes of packed files. the order must be guaranteed (by TCP).
ID, ranging [FileIdMin=1, FileIdMax=255], cannot promise matching, if multiple transfers are concurrent.
FileHdr1stLen=6 FileHdrRstLen=6

only one piece
bytes:0 1          2-5             6-         -
   2   ID  length of file name  file name  content

the first piece of many
bytes:0 1          2-5             6-         -
   1   ID  length of file name  file name  content
final piece
bytes:0 1   2-5      6-
   2   ID  reserved  content
other pieces
bytes:0 1   2-5      6-
   1   ID  reserved  content
*/

// IsDataFile checks whether received data a file (piece)
// It does not guarantee successful parsing into meaningful parts.
func IsDataFile(data []byte) (isData, isEnd bool) {
	if data == nil || len(data) < 1 {
		return false, false
	}
	switch data[0] {
	case 2:
		isEnd = true
		fallthrough
	case 1:
		if len(data) > FileHdr1stLen {
			return true, isEnd
		}
		eztools.Log("LESS data then header!", len(data), FileHdr1stLen, data[0])
	}

	return false, false
}

func CurrTime() string {
	return time.Now().Format("20060102-150405")
}

func GetAvailFileName(fn, addr string) (string, bool) {
	if _, err := os.Stat(fn); err != nil && os.IsNotExist(err) {
		return fn, true
	}
	// change IP address to valid file names
	addr = strings.ReplaceAll(addr, "[", "")
	addr = strings.ReplaceAll(addr, "]", "")
	addr = strings.ReplaceAll(addr, ":", ".")
	fn += "_" + addr + "_" + CurrTime()
	allTaken := true
	for nf, affix := fn, 1; affix < 100; affix += 1 {
		if _, err := os.Stat(nf); err != nil && os.IsNotExist(err) {
			fn = nf
			allTaken = false
			break
		}
		nf = fn + "_" + strconv.Itoa(affix)
	}
	if allTaken {
		return "", false
	}
	return fn, true
}

// BulkFile parses data into meaningful parts
// IsDataFile() must be called beforehand to ensure min length
func BulkFile(dir, affix string, data []byte) (fn string,
	first bool, cont []byte, end bool) {
	getFN := func() (fn string, fnEnd int) {
		fl := binary.LittleEndian.Uint32(data[2:FileHdr1stLen])
		fnEnd = int(FileHdr1stLen + fl)
		if len(data) < fnEnd {
			fnEnd = -1
		} else {
			fn = string(data[FileHdr1stLen:fnEnd])
		}
		return
	}
	switch data[0] {
	case 2:
		end = true
	}
	id := int(data[1])
	filePMLock.Lock()
	if filePieceMap == nil {
		filePieceMap = make(map[int]string)
	}
	fn, rec := filePieceMap[id]
	//eztools.Log("file piece", end, id, rec, fn)
	switch {
	case !rec:
		filePMLock.Unlock()
		first = true
		var fnEnd int
		fn, fnEnd = getFN()
		if fnEnd < 0 {
			if eztools.Debugging {
				eztools.Log("NO file name parsed!")
			}
			return
		}
		if len(dir) > 0 {
			fn = filepath.Join(dir, fn)
			if nfn, ok := GetAvailFileName(fn, affix); !ok {
				if eztools.Debugging {
					eztools.Log("NO available file names for",
						fn)
				}
				return
			} else {
				fn = nfn
			}
		}
		if !end {
			filePMLock.Lock()
			filePieceMap[id] = fn
			filePMLock.Unlock()
		}
		cont = data[fnEnd:]
		return
	case rec && end:
		delete(filePieceMap, id)
		filePMLock.Unlock()
	default:
		filePMLock.Unlock()
	}
	cont = data[FileHdrRstLen:]
	return
}

var (
	fileID       int
	fileIdLock   sync.Mutex
	filePieceMap map[int]string
	filePMLock   sync.Mutex
)

const (
	FileIdMax     = 255
	FileIdMin     = 1
	FileHdr1stLen = 6
	FileHdrRstLen = 6
)

func makeFileID() (ret int) {
	fileIdLock.Lock()
	switch fileID {
	case 0:
		rand.Seed(time.Now().UnixNano())
		fileID = rand.Intn(FileIdMax-FileIdMin) + FileIdMin
	case FileIdMax:
		fileID = FileIdMin
	default:
		fileID += 1
	}
	ret = fileID
	fileIdLock.Unlock()
	return ret
}

// prefix4File generates prefix for a file to send
// Return values:
//
//	prefix slice
//	errors from binary.Write()
func prefix4File(id int, indx uint32, end bool, fn string) ([]byte, error) {
	int2byte := func(i uint32, n int) (data []byte, err error) {
		buf := new(bytes.Buffer)
		err = binary.Write(buf, binary.LittleEndian, i)
		if err != nil {
			return
		}
		fl := buf.Bytes()
		if len(fl) > n {
			for i := n; i < len(fl); i++ {
				if fl[i] != 0 {
					//eztools.Log(fl)
					return nil, eztools.ErrOutOfBound
				}
			}
			fl = fl[:n]
		}
		data = make([]byte, n)
		copy(data, fl)
		return
	}
	// first byte
	ret := make([]byte, 1)
	if end {
		ret[0] = 2
	} else {
		ret[0] = 1
	}
	// ID
	idb, err := int2byte(uint32(id), 1)
	if err != nil {
		return nil, err
	}
	ret = append(ret, idb...)

	if indx == 0 && len(fn) > 0 {
		// file name
		fb, err := int2byte(uint32(len(fn)), 4)
		if err != nil {
			return nil, err
		}
		ret = append(ret, fb...)
		//eztools.Log("FL", ret)
		ret = append(ret, []byte(fn)...)
	} else {
		// reserved
		fb := make([]byte, 4)
		ret = append(ret, fb...)
	}
	return ret, nil
}

// Sz41stChunk returns max data size to transfer in first chunk
//
//	if input file name is too long, a valid one is returned
func Sz41stChunk(fnI string) (int, string) {
	ret := FlowRcvLen - FileHdr1stLen - len(fnI)
	if ret > 0 {
		return ret, fnI
	}
	fnO := EzcName
	return FlowRcvLen - FileHdr1stLen - len(fnO), fnO
}

// TryOnlyChunk try to read the file in one chunk
// Parameter: rdr is closed before returning
// Return values:
//
//	a possible long file name
//	read buffer
//	ErrOutOfBound if no valid data fits
//	ErrInvalidInput file larger than one chunk
func TryOnlyChunk(fnI string, rdr io.ReadCloser) (string, []byte, error) {
	defer rdr.Close()
	m, fnO := Sz41stChunk(fnI)
	if m <= 0 {
		return fnO, nil, eztools.ErrOutOfBound
	}
	buf := make([]byte, m+1)
	n, err := rdr.Read(buf)
	if err != nil && err == io.EOF {
		return fnO, buf[:n], nil
	}
	if err != nil {
		return fnO, nil, err
	}
	if n <= m {
		return fnO, buf[:n], nil
	}
	return fnO, nil, eztools.ErrInvalidInput
}

// SndFile sends a file without splitting.
// Parameter: rdr is closed before returning
// Return value:
//
//	ErrOutOfBound if file+prefix larger than FlowRcvLen
//	other error from os.Stat(), ioutil.ReadFile(), or prefix4File()
//	value from proc()
func SndFile(fn string, rdr io.ReadCloser, proc func([]byte) error) error {
	fn, buf, err := TryOnlyChunk(fn, rdr)
	if err != nil {
		return err
	}
	ret, err := prefix4File(makeFileID(), 0, true, fn)
	if err != nil {
		return err
	}
	return proc(append(ret, buf...))
}

// SplitFile
// Parameter: rdr is closed before returning
func SplitFile(fn string, rdr io.ReadCloser, proc func([]byte) error) error {
	readSz := func(indx uint32, name string) (uint32, string) {
		switch indx {
		case 0:
			sz, fnO := Sz41stChunk(fn)
			return uint32(sz), fnO
		default:
			return FlowRcvLen - FileHdrRstLen, name
		}
	}
	id := makeFileID()
	fun := func(indx uint32, name string, data []byte, done bool) error {
		ret, err := prefix4File(id, indx, done, name)
		if err != nil {
			return err
		}

		// content
		//eztools.Log("file sending", len(data), len(ret), ret)
		ret = append(ret, data...)
		return proc(ret)
	}
	return eztools.FileReaderByPiece(rdr, fn, readSz, fun)
}
