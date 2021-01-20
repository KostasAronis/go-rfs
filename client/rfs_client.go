package client

import (
	"errors"

	"github.com/KostasAronis/go-rfs/rfslib"
	"github.com/KostasAronis/go-rfs/tcp"
)

type RfsClient struct {
	tcpClient *tcp.Client
}

//CreateFile Creates a new empty RFS file with name fname.
//
// Can return the following errors:
// - DisconnectedError
// - FileExistsError
// - BadFilenameError
func (r *RfsClient) CreateFile(fname string) (err error) {
	tcpMsg := tcp.Msg{
		MSGType: tcp.CreateFile,
		Payload: map[string]interface{}{
			"FileName": fname,
		},
	}
	res := r.tcpClient.Send(&tcpMsg)
	if res.MSGType == tcp.Error {
		return errors.New(res.Payload.(string))
	}
	return nil
}

//ListFiles Returns a slice of strings containing filenames of all the
// existing files in RFS.
//
// Can return the following errors:
// - DisconnectedError
func (r *RfsClient) ListFiles() (fnames []string, err error) {
	tcpMsg := tcp.Msg{
		MSGType: tcp.ListFiles,
		Payload: map[string]interface{}{
			"FileName": "",
		},
	}
	res := r.tcpClient.Send(&tcpMsg)
	if res.MSGType == tcp.Error {

	}
	return nil, nil
}

//TotalRecs Returns the total number of records in a file with filename
// fname.
//
// Can return the following errors:
// - DisconnectedError
// - FileDoesNotExistError
func (r *RfsClient) TotalRecs(fname string) (numRecs uint16, err error) {
	tcpMsg := tcp.Msg{
		MSGType: tcp.TotalRecs,
		Payload: map[string]interface{}{
			"FileName": fname,
		},
	}
	res := r.tcpClient.Send(&tcpMsg)
	if res.MSGType == tcp.Error {

	}
	return 0, nil
}

//ReadRec Reads a record from file fname at position recordNum into
// memory pointed to by record. Returns a non-nil error if the
// read was unsuccessful. If a record at this index does not yet
// exist, this call must block until the record at this index
// exists, and then return the record.
//
// Can return the following errors:
// - DisconnectedError
// - FileDoesNotExistError
func (r *RfsClient) ReadRec(fname string, recordNum uint16, record *rfslib.Record) (err error) {
	tcpMsg := tcp.Msg{
		MSGType: tcp.ReadRec,
		Payload: map[string]interface{}{
			"FileName": fname,
			"Record":   recordNum,
		},
	}
	res := r.tcpClient.Send(&tcpMsg)
	if res.MSGType == tcp.Error {

	}
	return nil
}

//AppendRec Appends a new record to a file with name fname with the
// contents pointed to by record. Returns the position of the
// record that was just appended as recordNum. Returns a non-nil
// error if the operation was unsuccessful.
//
// Can return the following errors:
// - DisconnectedError
// - FileDoesNotExistError
// - FileMaxLenReachedError
func (r *RfsClient) AppendRec(fname string, record *rfslib.Record) (recordNum uint16, err error) {
	tcpMsg := tcp.Msg{
		MSGType: tcp.AppendRec,
		Payload: map[string]interface{}{
			"FileName": fname,
			"Record":   record,
		},
	}
	res := r.tcpClient.Send(&tcpMsg)
	if res.MSGType == tcp.Error {
		return 0, errors.New(res.Payload.(string))
	}
	return res.Payload.(uint16), nil
}
