package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var difficulty = 4

type NodeInfo struct {
	id     string
	path   string
	writer http.ResponseWriter
}

type Block struct {
	Index      int    // height of the block
	TimeStamp  int64  // transaction timestamp
	Data       string // transaction record
	Hash       string // SHA256 hash value of current node
	PrevHash   string // SHA256 hash value of previous node
	Nonce      int    // the number we are looking for in the PoW mining
	Difficulty int    // i.e. the count of zeros prefixing the hash value
}

//创建区块链
var blockchain []*Block

//创世区块
func genesisBlock() *Block {
	var geneBlock = Block{0, time.Now().Unix(), "", "", "", 0, difficulty}
	geneBlock.Hash = hex.EncodeToString(blockHash(geneBlock))
	return &geneBlock
}

func blockHash(block Block) []byte {
	re := strconv.Itoa(block.Index) + strconv.Itoa(int(block.TimeStamp)) + block.Data + block.PrevHash +
		strconv.Itoa(block.Nonce) + strconv.Itoa(block.Difficulty)
	h := sha256.New()
	h.Write([]byte(re))
	hashed := h.Sum(nil)
	return hashed
}

func isBlockValid(block Block) bool {
	prefix := strings.Repeat("0", block.Difficulty)
	return strings.HasPrefix(block.Hash, prefix)
}

func createNewBlock(lastBlock *Block, data string) *Block {
	var newBlock Block
	newBlock.Index = lastBlock.Index + 1
	newBlock.TimeStamp = time.Now().Unix()
	newBlock.Data = data
	newBlock.PrevHash = lastBlock.Hash
	newBlock.Difficulty = difficulty
	newBlock.Nonce = 0
	// begin mining - the difficulty depends on the count of zeros prefixing the hash value
	for {
		// calculate hash
		newBlock.Hash = hex.EncodeToString(blockHash(newBlock))
		if isBlockValid(newBlock) {
			// verify the block
			if verifyBlock(newBlock, *lastBlock) {
				fmt.Println("mining successful: ", newBlock.Hash)
				return &newBlock
			}
		}

		newBlock.Nonce++
	}
}

func verifyBlock(newblock Block, lastBlock Block) bool {
	if lastBlock.Index+1 != newblock.Index {
		return false
	}
	if newblock.PrevHash != lastBlock.Hash {
		return false
	}
	return true
}

var nodeTable = make(map[string]string)

func main() {

	userId := os.Args[1]
	fmt.Println("node " + userId)

	//./main Arsenal

	// initiate the addresses of the four countries
	nodeTable = map[string]string{
		"0": "localhost:1110",
		"1": "localhost:1111",
		"2": "localhost:1112",
		"3": "localhost:1113",
	}

	node := NodeInfo{userId, nodeTable[userId], nil}


	http.HandleFunc("/req", node.onRequest)

	// start up the server
	err := http.ListenAndServe(node.path, nil)
	if err != nil {
		fmt.Print(err)
	}

	var genBlock = genesisBlock()
	var newBlock *Block
	newBlock = genBlock
	for i := 0; i < 10; i++ {
		newBlock = createNewBlock(newBlock, fmt.Sprintf("new block %d", i))
		blockchain = append(blockchain, newBlock)
		fmt.Print("New block info: \n")
		fmt.Printf("height [%d], hash [%s], data [%s], nonce [%d], difficulty [%d].\n",
			newBlock.Index, newBlock.Hash, newBlock.Data, newBlock.Nonce, newBlock.Difficulty)
	}

	bytes, _ := json.MarshalIndent(blockchain, "", "  ")
	fmt.Println("========== Blockchain ==========")
	fmt.Println(string(bytes))
}
