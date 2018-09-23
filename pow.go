package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const (
	//当前挖矿的难度值
	difficult = 5
)

//实现区块链中的Pow算法

//简单的区块结构
type Block struct {
	//上一个区块
	PreHash []byte
	//时间戳
	Timestamp int64
	//交易信息
	Data []byte
	//当前区块的hash值
	Hash []byte
	//随机算
	Nonce int64
}

//创建创世区块
func gcreateGensisBlock() *Block {

	block := &Block{nil, time.Now().Unix(), []byte("gensis block"), nil, 0}
	//计算当前区块的Hash值
	block.setHash();
	return block
}
func (block *Block) setHash() {
	//拼接区块的信息
	//这里定义的规则是 ：前一个区块的Hash + 时间戳 + 交易信息 + Nonce .在取hash
	blockInfo := hex.EncodeToString(block.PreHash) + string(block.Timestamp) + hex.EncodeToString(block.Data) + string(block.Nonce)
	sha256 := sha256.New()
	sha256.Write([]byte(blockInfo))
	sum := sha256.Sum(nil)
	block.Hash = sum

}

//通过nonce值，计算hash值
func (block *Block) calHash(nonce int64) string {
	//拼接区块的信息
	//这里定义的规则是 ：前一个区块的Hash + 时间戳 + 交易信息 + Nonce .在取hash
	blockInfo := hex.EncodeToString(block.PreHash) + string(block.Timestamp) + hex.EncodeToString(block.Data) + string(nonce)
	sha256 := sha256.New()
	sha256.Write([]byte(blockInfo))
	sum := sha256.Sum(nil)
	return hex.EncodeToString(sum);
}

func getDiff() string {
	preHash := make([]byte, difficult)
	for i := 0; i < difficult; i++ {
		preHash[i] = '0';
	}
	return string(preHash)
}

//通过pow的方式模拟挖矿
func (b *Block) generateNextBlockByPow(data []byte) *Block {
	newBlock := &Block{b.PreHash, time.Now().Unix(), data, nil, 0}

	//开始挖矿，不断改变Nonce值，最终实现计算下来的hash的0的个数小于系统的难度值
	var nonce int64 = 1

	for {
		hash := b.calHash(nonce)
		if strings.HasPrefix(hash, getDiff()) {
			fmt.Println("挖矿成功");
			newBlock.Hash = []byte(hash)
			newBlock.Nonce = nonce
			return newBlock
		}
		nonce++
	}

}

//模拟挖矿
func main() {
	genesisBlock := gcreateGensisBlock()
	powBlock := genesisBlock.generateNextBlockByPow([]byte("new Blcok"))

	fmt.Println("data is ", string(powBlock.Data))
	fmt.Println("dataHex is ", hex.EncodeToString(powBlock.Data))
	fmt.Println("nonce is ", powBlock.Nonce)
}
