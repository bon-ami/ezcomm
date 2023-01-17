package main

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"testing"

	"gitee.com/bon-ami/eztools/v5"
	"gitlab.com/bon-ami/ezcomm"
)

var tstCntFile int

// TestFS uses TstRoot and TstMsgCount
func TestFS(t *testing.T) {
	ezcomm.Init4Tests(t)
	fs := httpFS(*ezcomm.TstRoot)
	tstCntFile = *ezcomm.TstMsgCount
	if tstCntFile == 0 {
		tstCntFile--
	}
	if err := tstFSRead(t, fs, ""); err != nil {
		t.Fatal(err)
	}
}

func tstFSRead(t *testing.T, fs httpFS, chld string) error {
	ff, err := fs.Open(chld)
	if err == nil {
		defer ff.Close()
		err = tstFFRead(t, fs, ff, chld)
	}
	if err != nil {
		t.Error(chld, "under", fs)
	}
	return err
}

func tstFFRead(t *testing.T, hfs httpFS, ff http.File, fn string) error {
	fi, err := ff.Stat()
	if err != nil {
		return err
	}
	cnt := fi.Size()
	isDir := fi.IsDir()
	if eztools.Verbose > 1 {
		t.Log(fi.Name(), "is dir =", isDir, "size =", cnt)
	}
	rdFile := func(fd []fs.FileInfo) error {
		for _, f1 := range fd {
			if tstCntFile == 0 {
				break
			}
			if err = tstFSRead(t, httpFS(filepath.Join(string(hfs), fn)),
				f1.Name()); err != nil {
				return err
			}
		}
		return nil
	}
	switch isDir {
	case true:
		if eztools.Verbose > 1 {
			t.Log("readdir -1")
		}
		fd, err := ff.Readdir(-1)
		if err != nil {
			return err
		}
		if err = rdFile(fd); err != nil {
			return err
		}
		cnt := len(fd)
		if eztools.Verbose > 1 {
			t.Log("readdir", cnt)
		}
		fd, err = ff.Readdir(cnt)
		if err != nil {
			return err
		}
		if cnt != len(fd) {
			t.Error(cnt, "<>", len(fd))
			return eztools.ErrOutOfBound
		}
		if err = rdFile(fd); err != nil {
			return err
		}
		if cnt < 1 {
			break
		}
		if eztools.Verbose > 1 {
			t.Log("readdir 0")
		}
		fd, err = ff.Readdir(0)
		if err != nil {
			return err
		}
		if cnt != len(fd) {
			t.Error(cnt, "<>", len(fd))
			return eztools.ErrOutOfBound
		}
		if err = rdFile(fd); err != nil {
			return err
		}
		if eztools.Verbose > 1 {
			t.Log("readdir 1/", cnt)
		}
		for ; cnt >= 0; cnt-- {
			fd, err = ff.Readdir(1)
			if len(Tst) > 0 {
				if eztools.Verbose > 1 {
					t.Log(Tst)
				}
			}
			Tst = nil
			if err != nil {
				return err
			}
			if len(fd) != 1 {
				return eztools.ErrInExistence
			}
			if err = rdFile(fd); err != nil {
				return err
			}
		}
		fd, err = ff.Readdir(1)
		if err == nil || len(fd) > 0 {
			t.Error("len fd", len(fd))
			return eztools.ErrOutOfBound
		}
	case false:
		//ff.Read()
		if tstCntFile == 0 {
			return nil
		}
		tstCntFile--
	}
	return nil
}
