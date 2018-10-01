package main

import (
	"crypto"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
var privK *rsa.PrivateKey
var pubK *rsa.PublicKey
var sessionK = "e3988cce1bdcd1db1b0a1313e598b12040d4e16f"

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

	privK, _ = rsa.GenerateKey(rand.Reader, 1024)
	pubK = &privK.PublicKey

	node := nodeInfo{userId, nodeTable[userId], nil}
	fmt.Println(node)

	//http协议的回调函数
	//http://localhost:1111/req?warTime=8888
	http.HandleFunc("/req", node.onRequest)
	http.HandleFunc("/prePrepare", node.prePrepare)
	http.HandleFunc("/prepare", node.prepare)
	http.HandleFunc("/commit", node.commit)

	// start up the server
	err := http.ListenAndServe(node.path, nil)
	if err != nil {
		fmt.Print(err)
	}

}

func generateRSASignature(msg string) ([]byte, []byte, *rsa.PSSOptions) {
	plaintext := []byte(msg)
	h := md5.New()
	h.Write(plaintext)

	hashed := h.Sum(nil)
	opts := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto, Hash: crypto.MD5}
	sign, _ := rsa.SignPSS(rand.Reader, privK, crypto.MD5, hashed, opts)
	return hashed, sign, opts
}

func generateMAC(data, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

func checkMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}

func (node *nodeInfo) onRequest(writer http.ResponseWriter, request *http.Request) {
	// receive and parse the params
	request.ParseForm()
	if len(request.Form["message"]) > 0 {
		node.writer = writer
		hashed, sign, opts := generateRSASignature(request.Form["message"][0])
		node.broadcastPrePrepare(sign, "/prePrepare", hashed, opts)
	}

}

// the broadcast in pre-prepare phase - primary 0 multicasts to 1,2,3
func (node *nodeInfo) broadcastPrePrepare(sign []byte, path string, hashed []byte, opts *rsa.PSSOptions) {
	e := rsa.VerifyPSS(pubK, crypto.MD5, hashed, []byte(sign), opts)

	if e != nil {
		fmt.Println("authentication failed")
		return
	}

	// loop through all replicas
	for nodeId, _url := range nodeTable {

		if nodeId == node.id {
			continue
		}
		message := "authenticating message from " + node.id + " to " + nodeId
		mac := generateMAC([]byte(message), []byte(sessionK))
		data := url.Values{"message": {message}, "mac": {string(mac)}, "nodeId": {node.id}}
		http.PostForm("http://"+_url+path, data)
	}

}

// the broadcast in prepare and commit phase
func (node *nodeInfo) broadcast(path string, message []byte, mac []byte) {
	for nodeId, _url := range nodeTable {

		if nodeId == node.id {
			continue
		}
		isMACEqual := checkMAC(message, []byte(mac), []byte(sessionK))
		if isMACEqual {
			haha := "http://" + _url + path
			data := url.Values{"message": {string(message)}, "mac": {string(mac)}, "nodeId": {node.id}}
			http.PostForm(haha, data)
		}
	}
}

// Primary 0 receives the onRequest from C and multicasts it to backups.
func (node *nodeInfo) prePrepare(writer http.ResponseWriter, request *http.Request) {
	if node.writer == nil {
		node.writer = writer
	}

	// receive and parse the params
	request.ParseForm()
	// distribute again
	if len(request.PostFormValue("message")) > 0 && len(request.PostFormValue("mac")) > 0 {
		// distribute to the other 3 nodes
		message := request.PostFormValue("message")
		mac := request.PostFormValue("mac")
		//nodeId := onRequest.Form["nodeId"][0]
		isMACEqual := checkMAC([]byte(message), []byte(mac), []byte(sessionK))
		if isMACEqual {
			newMsg := "message for prepare"
			newMAC := generateMAC([]byte(newMsg), []byte(sessionK))
			node.broadcast("/prepare", []byte(newMsg), newMAC)
		} else {
			fmt.Println("prePrepare: authentication failed")
		}
	}

}

// Replicas execute the onRequest and then re-broadcast the result to each other
func (node *nodeInfo) prepare(writer http.ResponseWriter, request *http.Request) {
	if node.writer == nil {
		node.writer = writer
	}
	request.ParseForm()
	if len(request.PostFormValue("message")) > 0 {
		fmt.Println(request.PostFormValue("message"))
	}
	if len(request.PostFormValue("mac")) > 0 {
		fmt.Println(request.PostFormValue("mac"))
	}
	if len(request.PostFormValue("nodeId")) > 0 {
		fmt.Println(request.PostFormValue("nodeId"))
	}

	node.authentication(request)
}

var authenticationMap = make(map[string]string)

// receive other nodes' info except that of self
func (node *nodeInfo) authentication(request *http.Request) {

	//接收参数
	request.ParseForm()

	if len(request.Form["nodeId"]) > 0 {
		authenticationMap[request.PostFormValue("nodeId")] = "ok"
	}

	// if the received msg count > node table count * 1/3
	if len(authenticationMap) > len(nodeTable)/3 {
		if len(request.PostFormValue("message")) > 0 && len(request.PostFormValue("mac")) > 0 {
			message := request.PostFormValue("message")
			mac := request.PostFormValue("mac")
			isMACEqual := checkMAC([]byte(message), []byte(mac), []byte(sessionK))
			if isMACEqual {
				// then PBFT consensus is achieved; commit the feedback to browser
				newMsg := "message for commit"
				newMAC := generateMAC([]byte(newMsg), []byte(sessionK))
				node.broadcast("/commit", []byte(newMsg), []byte(newMAC))
			} else {
				fmt.Println("prepare: authentication failed")
			}
		}
	}

}

// commit the feedback to the browser
func (node *nodeInfo) commit(writer http.ResponseWriter, request *http.Request) {
	if node.writer == nil {
		node.writer = writer
	}
	if len(request.PostFormValue("message")) > 0 && len(request.PostFormValue("mac")) > 0 {
		message := request.PostFormValue("message")
		mac := request.PostFormValue("mac")
		isMACEqual := checkMAC([]byte(message), []byte(mac), []byte(sessionK))
		if isMACEqual {
			// then PBFT consensus is achieved; commit the feedback to browser
			io.WriteString(node.writer, "authentication successful from node "+request.PostFormValue("nodeId")+"\n")
		} else {
			fmt.Println("commit: authentication failed")
		}
	}

}
