package main

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/KostasAronis/go-rfs/blockchain"
	"github.com/KostasAronis/go-rfs/serialization"
	"github.com/awalterschulze/gographviz"
)

var idToColor = map[string]string{
	"0": "black",
	"1": "green",
	"2": "blue",
	"3": "yellow",
	"4": "sienna",
}

//Takes as first cmd argument the name of a .bin file produced by a miner and creates a graphviz file of all blocks in the blockchain
func main() {
	var dataFileDir string
	if len(os.Args) > 1 {
		dataFileDir = os.Args[1]
	} else {
		panic("First argument should be path to .bin file")
	}
	files, err := ioutil.ReadDir(dataFileDir)
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".bin" {
			dataFilepath := path.Join(dataFileDir, f.Name())
			turnDatafileToGraphfile(dataFilepath)
		}
	}
}

func turnDatafileToGraphfile(dataFilepath string) {
	data, err := ioutil.ReadFile(dataFilepath)
	if err != nil {
		panic(err)
	}
	blockChain, err := serialization.DecodeToBlockTree(data)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile(dataFilepath+".dot", []byte(createGraph(blockChain).String()), 0666)
}

func createGraph(blockChain *blockchain.BlockTree) *gographviz.Graph {
	g := gographviz.NewGraph()
	if err := g.SetName("G"); err != nil {
		panic(err)
	}
	if err := g.SetDir(true); err != nil {
		panic(err)
	}
	for _, block := range blockChain.Blocks {
		h, err := block.ComputeHash()
		if err != nil {
			panic(err)
		}
		attrs := map[string]string{}
		//TODO: COLOR FOR OP / NOOP BLOCKS, COLOR FOR EXTERNAL / MINED BLOCKS
		if block.IsOp {
			attrs["color"] = "red"
		}
		attrs["fontcolor"] = idToColor[block.MinerID]

		g.AddNode("G", esc(h), attrs)
	}
	for _, block := range blockChain.Blocks {
		h, err := block.ComputeHash()
		if err != nil {
			panic(err)
		}
		if block.PrevHash != "" {
			g.AddEdge(esc(block.PrevHash), esc(h), true, nil)
		}
	}
	return g
}

func esc(h string) string {
	return "\"" + h + "\""
}
