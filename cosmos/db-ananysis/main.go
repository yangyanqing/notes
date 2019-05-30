package main

import (
	"fmt"
	"github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/syndtr/goleveldb/leveldb"
	cmn "github.com/tendermint/tendermint/libs/common"
)

type commitInfo struct {
	// Version
	Version int64

	// Store info for
	StoreInfos []storeInfo
}

type storeInfo struct {
	Name string
	Core storeCore
}

type storeCore struct {
	// StoreType StoreType
	CommitID CommitID
	// ... maybe add more state
}

type CommitID struct {
	Version int64
	Hash    []byte
}

type Node struct {
	key       []byte
	value     []byte
	version   int64
	height    int8
	size      int64
	hash      []byte
	leftHash  []byte
	leftNode  *Node
	rightHash []byte
	rightNode *Node
	persisted bool
}

func (node *Node) isLeaf() bool {
	return node.height == 0
}

const (
	RecTypeVersion = iota
	RecTypeBlock
	RecTypeKeeper
)

var (
	cdc = codec.New()
)

func parseRecord(key, value []byte) {
	switch getRecordType(key) {
	case RecTypeVersion:
		printVersion(key, value)
	case RecTypeKeeper:
		printKeeper(key, value)
	case RecTypeBlock:
		printBlock(key, value)
	}
}

//
// s/{height} block
// s/k:{keeper_name}/{type}{Hash}
// s/latest
//
func getRecordType(key []byte) int {
	switch key[2] {
	case 'k':
		return RecTypeKeeper
	case 'l':
		return RecTypeVersion
	default:
		return RecTypeBlock
	}
}

func printVersion(key, value []byte) {
	var latest int64
	if err := cdc.UnmarshalBinaryLengthPrefixed(value, &latest); err != nil {
		panic(err)
	}
	fmt.Printf("%s: %d\n", key, latest)
}

func printBlock(key, value []byte) {
	var info commitInfo
	if err := cdc.UnmarshalBinaryLengthPrefixed(value, &info); err != nil {
		panic(err)
	}
	fmt.Println("Block: ", string(key[2:]))

	for _, store := range info.StoreInfos {
		fmt.Printf("  %-15s: %s\n", store.Name, bytesToHex(store.Core.CommitID.Hash))
	}
}

func printKeeper(key, value []byte) {
	moduleName, nodeType, hash := parseKeeperKey(key)
	fmt.Printf("%s:%s:%s\n", moduleName, nodeType, hash)
	if nodeType == "node" {
		_ = printNode(value)
	} else if nodeType == "root" {
		fmt.Println("  root hash :", bytesToHex(value))
		fmt.Println("")
	}
}

func parseKeeperKey(key []byte) (moduleName, nodeType, hash string) {
	for i := 4; i < len(key); i++ {
		if key[i] == '/' {
			moduleName = string(key[4:i])
			switch string(key[i+1]) {
			case "r":
				nodeType = "root"
			case "n":
				nodeType = "node"
			case "o":
				nodeType = "oooo"
			}
			hash = bytesToHex(key[i+2:])
			return
		}
	}
	panic(fmt.Sprintf("Invalid keeper key: %s", bytesToHex(key)))
}

func bytesToHex(data []byte) string {
	result := ""
	for _, d := range data {
		result += fmt.Sprintf("%02X", d)
	}
	return result
}

func printNode(buf []byte) cmn.Error {

	// Read node header (height, size, version, key).
	height, n, cause := amino.DecodeInt8(buf)
	if cause != nil {
		return cmn.ErrorWrap(cause, "decoding node.height")
	}
	buf = buf[n:]

	size, n, cause := amino.DecodeVarint(buf)
	if cause != nil {
		return cmn.ErrorWrap(cause, "decoding node.size")
	}
	buf = buf[n:]

	ver, n, cause := amino.DecodeVarint(buf)
	if cause != nil {
		return cmn.ErrorWrap(cause, "decoding node.version")
	}
	buf = buf[n:]

	key, n, cause := amino.DecodeByteSlice(buf)
	if cause != nil {
		return cmn.ErrorWrap(cause, "decoding node.key")
	}
	buf = buf[n:]

	node := &Node{
		height:  height,
		size:    size,
		version: ver,
		key:     key,
	}

	// Read node body.

	if node.isLeaf() {
		val, _, cause := amino.DecodeByteSlice(buf)
		if cause != nil {
			return cmn.ErrorWrap(cause, "decoding node.value")
		}
		node.value = val
	} else { // Read children.
		leftHash, n, cause := amino.DecodeByteSlice(buf)
		if cause != nil {
			return cmn.ErrorWrap(cause, "deocding node.leftHash")
		}
		buf = buf[n:]

		rightHash, _, cause := amino.DecodeByteSlice(buf)
		if cause != nil {
			return cmn.ErrorWrap(cause, "decoding node.rightHash")
		}
		node.leftHash = leftHash
		node.rightHash = rightHash
	}
	var sValue string
	if node.value != nil && node.value[0] != 0 {
		sValue = string(node.value)
	} else {
		sValue = "n/a"
	}
	fmt.Println(
		fmt.Sprintf(""+
			"  height    : %d\n"+
			"  size      : %d\n"+
			"  version   : %d\n"+
			"  key       : %s\n"+
			"  left  hash: %s\n"+
			"  right hash: %s\n"+
			"  value     : %s\n",
			height, size, ver, string(key), bytesToHex(node.leftHash), bytesToHex(node.rightHash), sValue))
	return nil
}

func main() {
	db, err := leveldb.OpenFile("data/application.db", nil)
	if err != nil {
		panic(err)
	}

	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		parseRecord(key, value)
	}
}
