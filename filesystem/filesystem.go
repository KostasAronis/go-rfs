package filesystem

/*
	filesystem
	TODO: Either abstract the whole filesystem or inject an abstract storage layer
*/

import "github.com/KostasAronis/go-rfs/rfslib"

//FileSystem represents an in memory filesystem
type FileSystem struct {
	Files map[string]*File
}

func (f *FileSystem) Init() {
	f.Files = make(map[string]*File)
}

//ListFiles returns a slice of all filenames currently in the filesystem
func (f *FileSystem) ListFiles() []string {
	fileNames := make([]string, 0, len(f.Files))
	for k := range f.Files {
		fileNames = append(fileNames, k)
	}
	return fileNames
}

//AddFile adds a file without records (touch)
func (f *FileSystem) AddFile(fName string) (*File, error) {
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

func (f *FileSystem) TotalRecords(fName string) (int, error) {
	if file, exists := f.Files[fName]; !exists {
		return -1, rfslib.FileDoesNotExistError(fName)
	} else {
		records := file.GetRecords()
		return len(records), nil
	}
}

//AppendRecord adds the content given as a record to the file and returns the index of the added record
func (f *FileSystem) AppendRecord(fName string, record *rfslib.Record) (int, error) {
	file, exists := f.Files[fName]
	if !exists {
		return -1, rfslib.FileDoesNotExistError(fName)
	}
	idx := file.AddRecord(record)
	return idx, nil
}

func (f *FileSystem) ReadRecord(fName string, idx int) (*rfslib.Record, error) {
	file, exists := f.Files[fName]
	if !exists {
		return nil, rfslib.FileDoesNotExistError(fName)
	}
	return file.GetRecord(idx), nil
}
