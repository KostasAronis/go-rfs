package filesystem

import "github.com/KostasAronis/go-rfs/rfslib"

type File struct {
	Name    string
	Records []*rfslib.Record
}

//AddRecord adds the content given as a record to the file and returns the index of the added record
func (f *File) AddRecord(record *rfslib.Record) int {
	f.Records = append(f.Records, record)
	return len(f.Records) - 1
}

func (f *File) GetRecord(idx int) *rfslib.Record {
	return f.Records[idx]
}

func (f *File) GetRecords() []*rfslib.Record {
	return f.Records
}
