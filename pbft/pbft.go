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
type NodeInfo struct {
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

	// initiate the addresses of the four countries
	nodeTable = map[string]string{
		"0": "localhost:1110",
		"1": "localhost:1111",
		"2": "localhost:1112",
		"3": "localhost:1113",
	}

	privK, _ = rsa.GenerateKey(rand.Reader, 1024)
	pubK = &privK.PublicKey

	node := NodeInfo{userId, nodeTable[userId], nil}
	fmt.Println(node)

	http.HandleFunc("/req", node.onRequest)
	http.HandleFunc("/prePrepare", node.onPrePrepare)
	http.HandleFunc("/prepare", node.onPrepare)
	http.HandleFunc("/commit", node.onCommit)

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

// phase 2: broadcast the initial message from primary to backups
func (node *NodeInfo) onRequest(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("==========")
	fmt.Println("STAGE PRE-PREPARE")

	// receive and parse the params
	request.ParseForm()
	if len(request.Form["message"]) > 0 {
		node.writer = writer
		hashed, sign, opts := generateRSASignature(request.Form["message"][0])
		node.broadcastPrePrepare(sign, "/prePrepare", hashed, opts)
	} else {
		fmt.Println("pre-prepare: authentication failed")
	}

}

// the broadcast in pre-onPrepare phase - primary 0 multicasts to 1,2,3
func (node *NodeInfo) broadcastPrePrepare(sign []byte, path string, hashed []byte, opts *rsa.PSSOptions) {
	e := rsa.VerifyPSS(pubK, crypto.MD5, hashed, []byte(sign), opts)

	if e != nil {
		fmt.Println("pre-prepare: authentication failed")
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

// the broadcast in onPrepare and onCommit phase
func (node *NodeInfo) broadcast(path string, message []byte, mac []byte) {
	for nodeId, _url := range nodeTable {

		if nodeId == node.id {
			continue
		}
		isMACEqual := checkMAC(message, []byte(mac), []byte(sessionK))
		if isMACEqual {
			data := url.Values{"message": {string(message)}, "mac": {string(mac)}, "nodeId": {node.id}}
			http.PostForm("http://"+_url+path, data)
		}
	}
}

// phase 3: Primary 0 receives the onRequest from C and multicasts it to backups.
func (node *NodeInfo) onPrePrepare(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("==========")
	fmt.Println("STAGE PRE-PREPARE")

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
		nodeId := request.PostFormValue("nodeId")
		fmt.Println("message: " + message)
		fmt.Println("MAC: " + mac)
		fmt.Println("from node " + nodeId)
		isMACEqual := checkMAC([]byte(message), []byte(mac), []byte(sessionK))
		if isMACEqual {
			newMsg := "msg prepare"
			newMAC := generateMAC([]byte(newMsg), []byte(sessionK))
			node.broadcast("/prepare", []byte(newMsg), newMAC)
		} else {
			fmt.Println("prepare: authentication failed")
		}
	}

}

// phase 4: Replicas execute the onRequest and then re-broadcast the result to each other
func (node *NodeInfo) onPrepare(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("==========")
	fmt.Println("STAGE COMMIT")

	if node.writer == nil {
		node.writer = writer
	}
	request.ParseForm()
	if len(request.PostFormValue("message")) > 0 {
		fmt.Println("message: " + request.PostFormValue("message"))
	}
	if len(request.PostFormValue("mac")) > 0 {
		fmt.Println("MAC: " + request.PostFormValue("mac"))
	}
	if len(request.PostFormValue("nodeId")) > 0 {
		fmt.Println("from node " + request.PostFormValue("nodeId"))
	}

	node.authentication(request)
}

var authenticationMap = make(map[string]string)

// receive other nodes' info except that of self
func (node *NodeInfo) authentication(request *http.Request) {

	//接收参数
	request.ParseForm()

	if len(request.Form["nodeId"]) > 0 {
		authenticationMap[request.PostFormValue("nodeId")] = "authentication successful from node " + request.PostFormValue("nodeId") + "\n"
	}

	// if the received msg count > node table count * 1/3
	if len(authenticationMap) > len(nodeTable)/3 {
		if len(request.PostFormValue("message")) > 0 && len(request.PostFormValue("mac")) > 0 {
			message := request.PostFormValue("message")
			mac := request.PostFormValue("mac")
			isMACEqual := checkMAC([]byte(message), []byte(mac), []byte(sessionK))
			if isMACEqual {
				// then PBFT consensus is achieved; onCommit the feedback to browser
				newMsg := "msg commit"
				newMAC := generateMAC([]byte(newMsg), []byte(sessionK))
				node.broadcast("/commit", []byte(newMsg), []byte(newMAC))
			} else {
				fmt.Println("commit: authentication failed")
			}
		}
	}

}

// phase 5: reply the feedback to the browser
func (node *NodeInfo) onCommit(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("==========")
	fmt.Println("STAGE REPLY")
	if node.writer == nil {
		node.writer = writer
	}
	if len(request.PostFormValue("message")) > 0 && len(request.PostFormValue("mac")) > 0 {
		message := request.PostFormValue("message")
		mac := request.PostFormValue("mac")
		nodeId := request.PostFormValue("nodeId")
		fmt.Println("message: " + message)
		fmt.Println("MAC: " + mac)
		fmt.Println("from node " + nodeId)

		isMACEqual := checkMAC([]byte(message), []byte(mac), []byte(sessionK))
		if isMACEqual {
			// then PBFT consensus is achieved; onCommit the feedback to browser
			io.WriteString(node.writer, "authentication successful from node "+request.PostFormValue("nodeId")+"\n")
		} else {
			fmt.Println("onCommit: authentication failed")
		}
	}

}
