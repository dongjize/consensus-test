package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var difficulty int

type Block struct {
	Index      int // height of the block
	TimeStamp  int64
	Data       string // transaction record
	Hash       string
	PrevHash   string
	Nonce      int
	Difficulty int // i.e. the count of zeros prefixing the hash value
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
		//计算hash
		cuhash := hex.EncodeToString(blockHash(newBlock))
		newBlock.Hash = cuhash
		if isBlockValid(newBlock) {
			// verify the block
			if verifyBlock(newBlock, *lastBlock) {
				fmt.Println("mining successful: ", cuhash)
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

func main() {
	var genBlock = genesisBlock()
	var newBlock *Block
	newBlock = genBlock
	for i := 0; i < 6; i++ {
		newBlock = createNewBlock(newBlock, fmt.Sprintf("new block %d", i))
		blockchain = append(blockchain, newBlock)
		difficulty = i + 2
		fmt.Print("New block info: \n")
		fmt.Printf("height [%d], hash [%s], data [%s], nonce [%d], difficulty [%d].\n",
			newBlock.Index, newBlock.Hash, newBlock.Data, newBlock.Nonce, newBlock.Difficulty)
	}

	bytes, _ := json.MarshalIndent(blockchain, "", "  ")
	fmt.Println("========== Blockchain ==========")
	fmt.Println(string(bytes))
}
