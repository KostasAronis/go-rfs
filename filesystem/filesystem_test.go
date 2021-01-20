package filesystem_test

import (
	"testing"

	"github.com/KostasAronis/go-rfs/filesystem"
	"github.com/KostasAronis/go-rfs/rfslib"
)

func TestAddFile(t *testing.T) {
	fs := filesystem.FileSystem{}
	fs.Init()
	f1Name := "test1"
	f1, err := fs.AddFile(f1Name)
	if err != nil {
		t.Error(err)
	}
	if f1.Name != f1Name {
		t.Error("Incorrect file name")
	}
	if len(fs.Files) != 1 {
		t.Error("File not added")
	}
	_, err = fs.AddFile(f1Name)
	if err == nil {
		t.Error("Adding the same file should return error")
	}
	fErr, correctErrorType := err.(rfslib.FileExistsError)
	if !correctErrorType {
		t.Error("Adding the same file should return FileExistsError")
	}
	if fErr.Error() != rfslib.FileExistsError(f1Name).Error() {
		t.Error("Adding the same file should return FileExistsError with fileName provided as error message")
	}
}

func TestListFiles(t *testing.T) {
	fs := filesystem.FileSystem{}
	fs.Init()
	f1Name := "test1"
	fs.AddFile(f1Name)
	fileNames := fs.ListFiles()
	if len(fileNames) != 1 {
		t.Error("Filenames length should be 1")
	}
}

func TestAddRecord(t *testing.T) {
	fs := filesystem.FileSystem{}
	fs.Init()
	errName := "fileNotExists"
	errRec := rfslib.Record([512]byte{})
	copy(errRec[:], "ASDF")
	idx, err := fs.AppendRecord(errName, &errRec)
	if err == nil {
		t.Error("Appending record to a file that does not exist should return error")
	}
	if idx != -1 {
		t.Error("Appending record to a file that does not exist should return idx == -1")
	}
	recN, err := fs.TotalRecords(errName)
	if err == nil {
		t.Error("TotalRecords of a file that does not exist should return error")
	}
	if recN != -1 {
		t.Error("TotalRecords of a file that does not exist should return idx == -1")
	}
	f1Name := "test1"
	fs.AddFile(f1Name)

	rec := rfslib.Record([512]byte{})
	copy(errRec[:], "LoremIpsum")
	fs.AppendRecord(f1Name, &rec)
	recN, err = fs.TotalRecords(f1Name)
	if err != nil {
		t.Error("TotalRecords of a file that does not exist should return error")
	}
	if recN != 1 {
		t.Error("TotalRecords of a file that does not exist should return idx == -1")
	}
}
