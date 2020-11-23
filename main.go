package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
//	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)

// Block represents each 'item' in the blockchain
type Block struct {
	Index     int
	Timestamp string
	Public_Key string 
	Hash_val_file string
	Hash      string
	PrevHash  string
}

// Blockchain is a series of validated Blocks
var Blockchain []Block

// bcServer handles incoming concurrent Blocks
var bcServer chan []Block
var mutex = &sync.Mutex{}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	bcServer = make(chan []Block)

	// create genesis block
	t := time.Now()
	genesisBlock := Block{0, t.String(), "", "", "",""}
	spew.Dump(genesisBlock)
	Blockchain = append(Blockchain, genesisBlock)

	tcpPort := os.Getenv("ADDR")

	// start TCP and serve TCP server
	server, err := net.Listen("tcp", ":"+tcpPort)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("TCP  Server Listening on port :", tcpPort)
	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn)
	}

}

func handleConn(conn net.Conn) {

	defer conn.Close()
	io.WriteString(conn, "Enter a new Hash_val_file:")
	scanner := bufio.NewScanner(conn)
	

	// take in Hash_val_file from stdin and add it to blockchain after conducting necessary validation
	go func() {
		for scanner.Scan() {
			Hash_val_file := scanner.Text()
			//io.WriteString(conn, "\nEnter your Public_Key:")
			//scanner1 := bufio.NewScanner(conn)

			//for scanner1.Scan(){
			Public_Key := ""


			//if err != nil {
				//log.Printf("%v not a number: %v", scanner.Text(), err)
				//continue
			//}
			
			newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], Hash_val_file, Public_Key)
			if err != nil {
				log.Println(err)
				continue
			}
			if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
				newBlockchain := append(Blockchain, newBlock)
				replaceChain(newBlockchain)
			}

			bcServer <- Blockchain

			break;
			//io.WriteString(conn, "\nEnter a new Hash_val_file:")
			//io.WriteString(conn, "\nEnter your Public_Key:")
			//}
			
		}
	}()

	// simulate receiving broadcast
	go func() {
		for {
			time.Sleep(30 * time.Second)
			mutex.Lock()
			output, err := json.Marshal(Blockchain)
			if err != nil {
				log.Fatal(err)
			}
			mutex.Unlock()
			io.WriteString(conn, string(output))
		}
	}()

	for _ = range bcServer {
		spew.Dump(Blockchain)
	}

}

// make sure block is valid by checking index, and comparing the hash of the previous block
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

// make sure the chain we're checking is longer than the current blockchain
func replaceChain(newBlocks []Block) {
	mutex.Lock()
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
	mutex.Unlock()
}

// SHA256 hasing
func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + block.Hash_val_file + block.Public_Key + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// create a new block using previous block's hash
func generateBlock(oldBlock Block, Hash_val_file string, Public_Key string) (Block, error) {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Hash_val_file = Hash_val_file
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)
	newBlock.Public_Key=Public_Key

	return newBlock, nil
	
}