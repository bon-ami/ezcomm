package ezcomm

import (
	"bytes"
	"encoding/binary"
	"io/fs"
	"io/ioutil"
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
FileHdrLen=6

only one piece
bytes:0 1          2-5             6-         -
   2   ID  length of file name  file name  content

the first piece of many
bytes:0 1          2-5             6-         -
   1   ID  length of file name  file name  content
final piece
bytes:0 1   2-5      6-
   2   ID  offset  content
other pieces
bytes:0 1   2-5      6-
   1   ID  offset  content
*/

// IsDataFile checks whether received data a file (piece)
// It does not guarantee successful parsing into meaningful parts.
func IsDataFile(data []byte) bool {
	if data == nil || len(data) < 1 {
		return false
	}
	switch data[0] {
	case 1, 2:
		if len(data) > FileHdrLen {
			return true
		}
	}

	return false
}

func GetAvailFileName(fn, addr string) (string, bool) {
	if _, err := os.Stat(fn); err != nil && os.IsNotExist(err) {
		return fn, true
	}
	// change IP address to valid file names
	addr = strings.ReplaceAll(addr, "[", "")
	addr = strings.ReplaceAll(addr, "]", "")
	addr = strings.ReplaceAll(addr, ":", ".")
	fn += "_" + addr + "_" +
		time.Now().Format("20060102-150405")
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
	offset uint32, cont []byte, end bool) {
	getFN := func() (fn string, fnEnd int) {
		fl := binary.LittleEndian.Uint32(data[2:FileHdrLen])
		fnEnd = int(FileHdrLen + fl)
		if len(data) < fnEnd {
			fnEnd = -1
		} else {
			fn = string(data[FileHdrLen:fnEnd])
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
	switch {
	case !rec:
		filePMLock.Unlock()
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
	offset = binary.LittleEndian.Uint32(data[2:FileHdrLen])
	cont = data[FileHdrLen:]
	return
}

var (
	fileID       int
	fileIdLock   sync.Mutex
	filePieceMap map[int]string
	filePMLock   sync.Mutex
)

const (
	FileIdMax  = 255
	FileIdMin  = 1
	FileHdrLen = 6
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
//	prefix slice
//	errors from binary.Write()
func prefix4File(id int, end bool, fn string, offset uint32) ([]byte, error) {
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

	if offset == 0 && len(fn) > 0 {
		// file name
		fb, err := int2byte(uint32(len(fn)), 4)
		if err != nil {
			return nil, err
		}
		ret = append(ret, fb...)
		//eztools.Log("FL", ret)
		ret = append(ret, []byte(fn)...)
	} else {
		// offset
		fb, err := int2byte(offset, 4)
		if err != nil {
			return nil, err
		}
		ret = append(ret, fb...)
	}
	return ret, nil
}

// SndFile sends a file without splitting.
// Return value:
//	ErrOutOfBound if file+prefix larger than FlowRcvLen
//	other error from os.Stat(), ioutil.ReadFile(), or prefix4File()
//	value from proc()
func SndFile(fn string, proc func([]byte) error) error {
	fi, err := os.Stat(fn)
	if err != nil {
		return err
	}
	if fi.Size() > FlowRcvLen {
		return eztools.ErrOutOfBound
	}
	buf, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}
	ret, err := prefix4File(makeFileID(), true, filepath.Base(fn), 0)
	if err != nil {
		return err
	}
	if fi.Size()+int64(len(ret)) > FlowRcvLen {
		return eztools.ErrOutOfBound
	}
	return proc(append(ret, buf...))
}

func SplitFile(fn string, proc func([]byte) error) error {
	readSz := func(indx uint32, inf fs.FileInfo, done uint32) uint32 {
		var pad, sz uint32
		switch indx {
		case 0:
			nm := inf.Name()
			pad = FileHdrLen + (uint32)(len(nm))
			sz = uint32(inf.Size()) + pad
		default:
			sz = uint32(inf.Size()) - done
			pad = FileHdrLen
		}
		//eztools.Log("size for pad", indx, pad, done, sz)
		if sz <= FlowRcvLen {
			return sz
		}
		return uint32(FlowRcvLen - pad)
	}
	id := makeFileID()
	fun := func(indx uint32, inf fs.FileInfo,
		offset uint32, data []byte) error {
		var end bool
		if offset+uint32(len(data)) >= uint32(inf.Size()) {
			end = true
		}
		ret, err := prefix4File(id, end, inf.Name(), offset)
		if err != nil {
			return err
		}

		// content
		ret = append(ret, data...)
		//eztools.Log("file sending", len(data), len(ret), ret)
		return proc(ret)
	}
	return eztools.FileReadByPiece(fn, readSz, fun)
}
