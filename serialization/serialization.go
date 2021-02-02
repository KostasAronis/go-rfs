package serialization

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"io/ioutil"
	"os"

	"github.com/KostasAronis/go-rfs/blockchain"
)

func EncodeToBytes(p interface{}) ([]byte, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(p)
	if err != nil {
		return nil, err
	}
	//fmt.Println("uncompressed size (bytes): ", len(buf.Bytes()))
	return buf.Bytes(), nil
}

func Compress(s []byte) ([]byte, error) {
	zipbuf := bytes.Buffer{}
	zipped := gzip.NewWriter(&zipbuf)
	_, err := zipped.Write(s)
	if err != nil {
		return nil, err
	}
	err = zipped.Close()
	if err != nil {
		return nil, err
	}
	//fmt.Println("compressed size (bytes): ", len(zipbuf.Bytes()))
	return zipbuf.Bytes(), nil
}

func Decompress(s []byte) ([]byte, error) {
	rdr, _ := gzip.NewReader(bytes.NewReader(s))
	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		return nil, err
	}
	err = rdr.Close()
	if err != nil {
		return nil, err
	}
	//fmt.Println("uncompressed size (bytes): ", len(data))
	return data, nil
}

func DecodeToBlockTree(s []byte) (*blockchain.BlockTree, error) {
	p := blockchain.BlockTree{}
	dec := gob.NewDecoder(bytes.NewReader(s))
	err := dec.Decode(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func WriteToFile(s []byte, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	_, err = f.Write(s)
	if err != nil {
		return err
	}
	return nil
}

func ReadFromFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return data, nil
}
