package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/KostasAronis/go-rfs/rfslib"
	"github.com/KostasAronis/go-rfs/tcp"
)

func helpStr() string {
	return `
Client to the go-rfs blockchain filesystem.

Usage:
		./client <ACTION> required-arguments [optional-arguments]
The possible actions are:
		ls	[-a]	 	:lists the files in the global namespace. The optional -a argument can be used to also list the number of records in each file.
		cat	fname	 	:output all of the records in fname to stdout.
		tail	k fname	 	:outputs the last k records in fname to stdout.
		head	k fname	 	:outputs the first k records in fname to stdout.
		append	fname str	:appends a new string to fname.
		touch	fname	 	:creates a blank file fname.
`
}

func help() {
	log.Println(helpStr())
}

// TODO: flesh out, currently just for testing
func main() {
	if len(os.Args) > 1 {
		err := handleAction(os.Args...)
		if err != nil {
			panic(err)
		}
	} else {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("-> ")
			text, _ := reader.ReadString('\n')
			text = strings.Replace(text, "\n", "", -1)
			text = strings.Replace(text, "\r", "", -1)
			if text == "exit" {
				return
			}
			arguments := strings.Split(text, " ")
			err := handleAction(append([]string{""}, arguments...)...)
			if err != nil {
				log.Println(err.Error())
			}
		}
	}
}

func handleAction(args ...string) error {
	action := args[1]
	switch action {
	case "save":
		err := saveAndExit()
		if err != nil {
			return err
		}
	case "ls":
		err := listFiles()
		if err != nil {
			return err
		}
	case "cat":
		filename := args[2]
		totalRecs, err := recCount(filename)
		if err != nil {
			return err
		}
		err = getRecords(filename, makeRange(0, totalRecs, totalRecs))
		if err != nil {
			return err
		}
	case "tail":
		filename := args[2]
		n := 5
		if len(args) > 3 {
			nArg, err := strconv.Atoi(args[3])
			if err != nil {
				help()
				return nil
			}
			n = nArg
		}
		totalRecs, err := recCount(filename)
		if err != nil {
			return err
		}
		err = getRecords(filename, makeRange(totalRecs-n, totalRecs, totalRecs))
		if err != nil {
			return err
		}
	case "head":
		filename := args[2]
		n := 5
		if len(args) > 3 {
			nArg, err := strconv.Atoi(args[3])
			if err != nil {
				help()
				return nil
			}
			n = nArg
		}
		totalRecs, err := recCount(filename)
		if err != nil {
			return err
		}
		getRecords(filename, makeRange(0, n, totalRecs))
	case "append":
		filename := args[2]
		record := args[3]
		err := appendRec(filename, record)
		if err != nil {
			return err
		}
	case "touch":
		filename := args[2]
		err := touch(filename)
		if err != nil {
			return err
		}
	default:
		help()
		return nil
	}
	return nil
}

func saveAndExit() error {
	msg := tcp.Msg{
		MSGType: tcp.StoreAndStop,
		Payload: map[string]interface{}{},
	}
	res, err := send(&msg)
	if err != nil {
		return err
	}
	resMap := res.(map[string]interface{})
	filename := resMap["Filename"]
	log.Printf("Stored in file: %s", filename)
	return nil
}
func getRecords(filename string, indexes []int) error {
	msg := tcp.Msg{
		MSGType: tcp.ReadRec,
		Payload: map[string]interface{}{
			"Filename": filename,
			"Record":   indexes,
		},
	}
	res, err := send(&msg)
	if err != nil {
		return err
	}
	recArr := res.([]interface{})
	for _, rec := range recArr {
		r := rfslib.Record{}
		r.FromFloatArrayInterface(rec)
		log.Println(r.ToString())
	}
	return nil
}

func recCount(filename string) (int, error) {
	msg := tcp.Msg{
		MSGType: tcp.TotalRecs,
		Payload: map[string]interface{}{
			"Filename": filename,
		},
	}
	res, err := send(&msg)
	if err != nil {
		return 0, err
	}
	return int(res.(float64)), nil
}

func listFiles() error {
	msg := tcp.Msg{
		MSGType: tcp.ListFiles,
	}
	res, err := send(&msg)
	if err != nil {
		return err
	}
	log.Println(res)
	return nil
}

func appendRec(filename string, record string) error {
	msg := tcp.Msg{
		MSGType: tcp.AppendRec,
		Payload: map[string]interface{}{
			"Filename": filename,
			"Record":   strToRec(record),
		},
	}
	res, err := send(&msg)
	if err != nil {
		return err
	}
	log.Println(res)
	return nil
}
func touch(filename string) error {
	msg := tcp.Msg{
		MSGType: tcp.CreateFile,
		Payload: map[string]interface{}{
			"Filename": filename,
			"Record":   strToRec(""),
		},
	}
	res, err := send(&msg)
	if err != nil {
		return err
	}
	log.Println(res)
	return nil
}

func send(msg *tcp.Msg) (interface{}, error) {
	c := tcp.Client{
		ID:         "c_1",
		Address:    "who cares",
		TargetAddr: ":8001",
		TargetID:   "1",
	}
	res := c.Send(msg, "")
	if res.MSGType == tcp.Error {
		return nil, fmt.Errorf(res.Payload.(string))
	}
	return res.Payload, nil
}

func strToRec(str string) *rfslib.Record {
	r := rfslib.Record{}
	r.FromString(str)
	//copy(r[:], str[:])
	return &r
}

func makeRange(min, max, total int) []int {
	if min < 0 {
		min = 0
	}
	if max > total {
		max = total
	}
	a := make([]int, max-min)
	for i := range a {
		a[i] = min + i
	}
	return a
}
