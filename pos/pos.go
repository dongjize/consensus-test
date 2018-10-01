package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type PBlock struct {
	Index     int    // height of the block
	Data      string // (transaction) data
	PreHash   string // SHA256 hash value of previous node
	Hash      string // SHA256 hash value of current node
	Timestamp string // transaction timestamp
	Validator *PNode // the node that validates this block
}

func genesisBlock() PBlock {
	var genesisBlock = PBlock{0, "Genesis block", "", "", time.Now().String(), &PNode{0, 0, "dd"}}
	genesisBlock.Hash = hex.EncodeToString(blockHash(&genesisBlock))
	return genesisBlock
}

func blockHash(block *PBlock) []byte {
	record := strconv.Itoa(block.Index) + block.Data + block.PreHash + block.Timestamp + block.Validator.Address
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hashed
}

type PNode struct {
	Tokens  int    // amount of tokens in stock
	Days    int    // time of stock
	Address string // node address
}

//创建5个节点
//算法的实现要满足 持币越多的节点越容易出块
var nodes = make([]PNode, 5)

//存放节点的地址
var addr = make([]*PNode, 15)

func InitNodes() {
	nodes[0] = PNode{1, 1, "0x12341"}
	nodes[1] = PNode{2, 1, "0x12342"}
	nodes[2] = PNode{3, 1, "0x12343"}
	nodes[3] = PNode{4, 1, "0x12344"}
	nodes[4] = PNode{5, 1, "0x12345"}
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

//采用Pos共识算法进行挖矿
func CreateNewBlock(lastBlock *PBlock, data string) PBlock {
	var newBlock PBlock
	newBlock.Index = lastBlock.Index + 1
	newBlock.Timestamp = time.Now().String()
	newBlock.PreHash = lastBlock.Hash
	newBlock.Data = data
	//通过pos计算由那个村民挖矿
	//设置随机种子
	time.Sleep(100000000)
	rand.Seed(time.Now().Unix())
	//[0,15)产生0-15的随机值
	var rd = rand.Intn(15)
	//选出挖矿的旷工
	node := addr[rd]
	fmt.Printf("Now node %s produce block by pos algorithm.\n", node.Address)
	//设置当前区块挖矿地址为旷工
	newBlock.Validator = node
	//简单模拟 挖矿所得奖励
	node.Tokens += 1
	newBlock.Hash = hex.EncodeToString(blockHash(&newBlock))
	return newBlock
}
func main() {
	InitNodes()
	var genesisBlock = genesisBlock()
	for i := 0; i < 100; i++ {
		var newBlock = CreateNewBlock(&genesisBlock, "new block")
		fmt.Print("New block info: \n")
		fmt.Printf("Hash: %s, Coinbase: %s.\n", newBlock.Hash, newBlock.Validator.Address)
	}
}
