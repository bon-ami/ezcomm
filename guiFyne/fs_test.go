package main

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"testing"

	"gitee.com/bon-ami/eztools/v5"
)

var tstCntFile int

const TstCntFile = -1

func TestFS(t *testing.T) {
	/*tstFile := "testing"
	t.Log(storage.ParseURI("."))
	uri := storage.NewFileURI(filepath.Join(".", tstFile))
	t.Log("creating", uri.String())
	wr, err := storage.Writer(uri)
	if err != nil {
		t.Fatal(err)
	}
	wr.Write([]byte{0, 1, 2})
	wr.Close()
	t.Log("created", uri.String())*/
	fs := httpFS(".")
	tstCntFile = TstCntFile
	if err := tstFSRead(t, fs, ""); err != nil {
		t.Fatal(err)
	}
	/*tstFSRead(t, fs, tstFile)
	err = storage.Delete(uri)
	if err != nil {
		t.Fatal(err)
	}*/
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
	t.Log(fi.Name(), "is dir =", isDir, "size =", cnt)
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
		t.Log("readdir -1")
		fd, err := ff.Readdir(-1)
		if err != nil {
			return err
		}
		if err = rdFile(fd); err != nil {
			return err
		}
		cnt := len(fd)
		t.Log("readdir", cnt)
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
		t.Log("readdir 0")
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
		t.Log("readdir 1/", cnt)
		for ; cnt >= 0; cnt-- {
			fd, err = ff.Readdir(1)
			if len(Tst) > 0 {
				t.Log(Tst)
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
