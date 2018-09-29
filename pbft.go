package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// declare node info, representing small countries
type nodeInfo struct {
	id     string
	path   string
	writer http.ResponseWriter
}

// The address map of the 4 countries
var nodeTable = make(map[string]string)

func main() {

	userId := os.Args[1]
	fmt.Println(userId)

	//./main Arsenal

	// initiate the addresses of the four countries
	nodeTable = map[string]string{
		"Arsenal":   "localhost:1111",
		"Chelsea":   "localhost:1112",
		"Liverpool": "localhost:1113",
		"ManUtd":    "localhost:1114",
	}

	node := nodeInfo{userId, nodeTable[userId], nil}
	fmt.Println(node)

	//http协议的回调函数
	//http://localhost:1111/req?warTime=8888
	http.HandleFunc("/req", node.request)
	http.HandleFunc("/prePrepare", node.prePrepare)
	http.HandleFunc("/prepare", node.prepare)
	http.HandleFunc("/commit", node.commit)

	// start up the server
	err := http.ListenAndServe(node.path, nil)
	if err != nil {
		fmt.Print(err)
	}

}

func (node *nodeInfo) request(writer http.ResponseWriter, request *http.Request) {
	// receive and parse the params
	request.ParseForm()
	//如果有参数值，则继续处理
	if len(request.Form["warTime"]) > 0 {
		node.writer = writer
		//激活主节点后，广播给其他节点,通过Arsenal向其他节点做广播
		node.broadcast(request.Form["warTime"][0], "/prePrepare")
	}

}

//由主节点向其他节点做广播
func (node *nodeInfo) broadcast(msg string, path string) {
	//遍历所有的国家
	for nodeId, url := range nodeTable {

		if nodeId == node.id {
			continue
		}
		//调用Get请求
		//http.Get("http://localhost:1112/prePrepare?warTime=8888&nodeId=Arsenal")
		http.Get("http://" + url + path + "?warTime=" + msg + "&nodeId=" + node.id)
	}

}

func (node *nodeInfo) prePrepare(writer http.ResponseWriter, request *http.Request) {
	if node.writer == nil {
		node.writer = writer
	}

	// receive and parse the params
	request.ParseForm()
	//fmt.Println("hello world")
	// distribute again
	if len(request.Form["warTime"]) > 0 {
		// distribute to the other 3 nodes
		node.broadcast(request.Form["warTime"][0], "/prepare")
	}

}

func (node *nodeInfo) prepare(writer http.ResponseWriter, request *http.Request) {

	request.ParseForm()
	//调用验证
	if len(request.Form["warTime"]) > 0 {
		fmt.Println(request.Form["warTime"][0])
	}
	if len(request.Form["nodeId"]) > 0 {
		fmt.Println(request.Form["nodeId"][0])
	}

	node.authentication(request)
}

var authenticationSuccess = true
var authenticationMap = make(map[string]string)

// receive other nodes' info except that of self
func (node *nodeInfo) authentication(request *http.Request) {

	//接收参数
	request.ParseForm()

	if authenticationSuccess != false {
		if len(request.Form["nodeId"]) > 0 {
			authenticationMap[request.Form["nodeId"][0]] = "ok"
		}
	}

	// if the received msg count > node table count * 1/3
	if len(authenticationMap) > len(nodeTable)/3 {
		// then PBFT consensus is achieved; commit the feedback to browser
		node.broadcast(request.Form["warTime"][0], "/commit")

	}
}

func (node *nodeInfo) commit(writer http.ResponseWriter, request *http.Request) {

	// commit the feedback to the browser
	io.WriteString(node.writer, "ok")

}
