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
	Index     int
	Timestamp string
	PrevHash  string
	Hash      string
	Data      []byte
	Delegate  *Node // the miner
}

func GenesisBlock() Block {
	gene := Block{0, time.Now().String(), "", "", []byte("genesis block"), nil}
	gene.Hash = string(calculateHash(gene))
	return Block{}
}

// generate the hash of a block
func calculateHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + block.PrevHash + hex.EncodeToString(block.Data)
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

//节点类型
type Node struct {
	Name  string //节点名称
	Votes int    // 被选举的票数
}

func (node *Node) GenerateNewBlock(lastBlock Block, data []byte) Block {
	var newBlock = Block{lastBlock.Index + 1, time.Now().String(), lastBlock.Hash, "", data, nil}
	newBlock.Hash = calculateHash(newBlock)
	newBlock.Delegate = node
	return newBlock
}

//创建节点
var NodeArr = make([]Node, 100)

func CreateNode() {
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("NODE %d num.", i+1)
		NodeArr[i] = Node{name, 0}
	}
}

//简单模拟投票
func Vote() {
	for i := 0; i < 100; i++ {
		rand.Seed(time.Now().UnixNano())
		time.Sleep(100000)
		vote := rand.Intn(10000) + 1
		NodeArr[i].Votes = vote
		fmt.Printf("Node [%d] votes is [%d].\n", i, vote)
	}
}

//elect the 21 nodes with most votes
func SortNodes() []Node {
	n := NodeArr
	for i := 0; i < len(n); i++ {
		for j := 0; j < len(n)-1; j++ {
			if n[j].Votes < n[j+1].Votes {
				n[j], n[j+1] = n[j+1], n[j]
			}
		}
	}
	return n[:21]
}
func main() {
	CreateNode()
	fmt.Print("###### Create node list: \n")
	fmt.Println(NodeArr)
	fmt.Print("###### Vote node: \n")
	Vote()
	nodes := SortNodes()
	fmt.Print("###### Get super node: \n")
	fmt.Println(nodes)
	// create the genesis block
	genesisBlock := GenesisBlock()
	newBlock := genesisBlock

	blockchain = append(blockchain, genesisBlock)

	fmt.Print("###### Begin producing block: \n")
	for i := 0; i < len(nodes); i++ {
		fmt.Printf("Node [%s] genenrates block with votes %d.\n", nodes[i].Name, nodes[i].Votes)
		newBlock = nodes[i].GenerateNewBlock(newBlock, []byte(fmt.Sprintf("new block %d", i)))

		if isBlockValid(newBlock, blockchain[len(blockchain)-1]) {
			blockchain = append(blockchain, newBlock)
		}
		str0, _ := json.MarshalIndent(newBlock, "", " ")
		fmt.Printf("%s\n", str0)
	}
}
