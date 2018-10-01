package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

var blockchain []Block

type Block struct {
	Index     int    // height of the block
	Data      string // (transaction) data
	PrevHash  string // SHA256 hash value of previous node
	Hash      string // SHA256 hash value of current node
	Timestamp string // transaction timestamp
	Validator *Node  // the node that validates this block
}

func genesisBlock() Block {
	var genesisBlock = Block{0, "Genesis block", "", "", time.Now().String(), &Node{0, 0, "dd"}}
	genesisBlock.Hash = calculateHash(genesisBlock)
	return genesisBlock
}

func calculateHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Data + block.PrevHash + block.Timestamp + block.Validator.Address
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

type Node struct {
	Tokens  int    // amount of tokens in stock
	Days    int    // time of stock
	Address string // node address
}

// create 5 nodes
// the more the node stakes, the easier it will win
var nodes = make([]Node, 5)

var addr = make([]*Node, 15)

func initNodes() {
	nodes[0] = Node{1, 1, "0x12341"}
	nodes[1] = Node{2, 1, "0x12342"}
	nodes[2] = Node{3, 1, "0x12343"}
	nodes[3] = Node{4, 1, "0x12344"}
	nodes[4] = Node{5, 1, "0x12345"}
	cnt := 0
	for i := 0; i < 5; i++ {
		for j := 0; j < nodes[i].Tokens*nodes[i].Days; j++ {
			addr[cnt] = &nodes[i]
			cnt++
		}
	}
	fmt.Print("Node list with [Tokens, Days, Address]:\n")
	fmt.Printf("%v \n", nodes)
	fmt.Print("Producer node set is: \n")
	for i := 0; i < len(addr); i++ {
		fmt.Printf("%v ", addr[i].Address)
	}
	fmt.Print("\n")
}

func createNewBlock(lastBlock Block, data string) Block {
	var newBlock Block
	newBlock.Index = lastBlock.Index + 1
	newBlock.Timestamp = time.Now().String()
	newBlock.PrevHash = lastBlock.Hash
	newBlock.Data = data
	time.Sleep(100000000)
	randMaker := rand.New(rand.NewSource(time.Now().UnixNano()))
	var rd = randMaker.Intn(15)
	fmt.Print(rd)
	node := addr[rd]
	fmt.Println()
	fmt.Printf("Node %s adds a block.\n", node.Address)
	newBlock.Validator = node
	// the winner can get one token as reward
	node.Tokens += 1
	newBlock.Hash = calculateHash(newBlock)
	return newBlock
}

func main() {
	initNodes()
	var genesisBlock = genesisBlock()
	blockchain = append(blockchain, genesisBlock)
	for i := 0; i < 20; i++ {
		var newBlock = createNewBlock(blockchain[len(blockchain)-1], "new block")
		if isBlockValid(newBlock, blockchain[len(blockchain)-1]) {
			blockchain = append(blockchain, newBlock)
		}
		str0, _ := json.MarshalIndent(newBlock, "", " ")
		fmt.Printf("%s\n", str0)
	}

}
