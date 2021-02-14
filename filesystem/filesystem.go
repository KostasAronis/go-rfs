package filesystem

/*
	filesystem
	TODO:
	1) Either abstract the whole filesystem or inject an abstract storage layer
	2) Does it need a sync Map instead o plain map
*/

import (
	"sync"

	"github.com/KostasAronis/go-rfs/rfslib"
)

//FileSystem represents an in memory filesystem
type FileSystem struct {
	Files map[string]*File
	m     sync.RWMutex
}

func (f *FileSystem) Init() {
	f.Files = make(map[string]*File)
	f.m = sync.RWMutex{}
}

func (f *FileSystem) Clone() *FileSystem {
	f.m.RLock()
	defer f.m.RUnlock()
	clone := FileSystem{
		Files: make(map[string]*File),
	}
	for k, v := range f.Files {
		cloneFile := &File{
			Name:    v.Name,
			Records: []*rfslib.Record{},
		}
		for _, r := range v.Records {
			cloneR := rfslib.Record(*r)
			cloneFile.Records = append(cloneFile.Records, &cloneR)
		}
		clone.Files[k] = cloneFile
	}
	return &clone
}

//AddFile adds a file without records (touch)
func (f *FileSystem) AddFile(fName string) (*File, error) {
	f.m.Lock()
	defer f.m.Unlock()
	if _, exists := f.Files[fName]; exists {
		return nil, rfslib.FileExistsError(fName)
	}
	file := File{
		Name:    fName,
		Records: []*rfslib.Record{},
	}
	f.Files[fName] = &file
	return &file, nil
}

//AppendRecord adds the content given as a record to the file and returns the index of the added record
func (f *FileSystem) AppendRecord(fName string, record *rfslib.Record) (int, error) {
	f.m.Lock()
	defer f.m.Unlock()
	file, exists := f.Files[fName]
	if !exists {
		return -1, rfslib.FileDoesNotExistError(fName)
	}
	idx := file.AddRecord(record)
	return idx, nil
}

//ListFiles returns a slice of all filenames currently in the filesystem
func (f *FileSystem) ListFiles() []string {
	f.m.RLock()
	defer f.m.RUnlock()
	fileNames := make([]string, 0, len(f.Files))
	for k := range f.Files {
		fileNames = append(fileNames, k)
	}
	return fileNames
}

func (f *FileSystem) TotalRecords(fName string) (int, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	if file, exists := f.Files[fName]; !exists {
		return -1, rfslib.FileDoesNotExistError(fName)
	} else {
		records := file.GetRecords()
		return len(records), nil
	}
}

func (f *FileSystem) ReadRecord(fName string, idx int) (*rfslib.Record, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	file, exists := f.Files[fName]
	if !exists {
		return nil, rfslib.FileDoesNotExistError(fName)
	}
	return file.GetRecord(idx), nil
}
