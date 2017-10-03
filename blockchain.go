package myblockchain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type BlockchainService interface {
	RegisterNode(address string) bool
	ValidChain(chain Blockchain) bool
	ResolveConflicts() bool
	NewBlock(proof int64, previousHash string) Block
	NewTransaction(tx Transaction) int64
	LastBlock() Block
	ProofOfWork(lastProof int64)
	VerifyProof(lastProof, proof int64) bool
}

type Block struct {
	Index        int64         `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	Proof        int64         `json:"proof"`
	PreviousHash string        `json:"previous_hash"`
}

type Transaction struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    int64  `json:"amount"`
}

type Blockchain struct {
	chain        []Block
	transactions []Transaction
	nodes        StringSet
}

func (bc *Blockchain) NewBlock(proof int64, previousHash string) Block {
	prevHash := previousHash
	if previousHash == "" {
		prevBlock := bc.chain[len(bc.chain)-1]
		prevHash = computeHashForBlock(prevBlock)
	}

	newBlock := Block{
		Index:        int64(len(bc.chain) + 1),
		Timestamp:    time.Now().UnixNano(),
		Transactions: bc.transactions,
		Proof:        proof,
		PreviousHash: prevHash,
	}

	bc.transactions = nil
	bc.chain = append(bc.chain, newBlock)
	return newBlock
}

func (bc *Blockchain) NewTransaction(tx Transaction) int64 {
	bc.transactions = append(bc.transactions, tx)
	return bc.LastBlock().Index + 1
}

func (bc *Blockchain) LastBlock() Block {
	return bc.chain[len(bc.chain)-1]
}

func (bc *Blockchain) ProofOfWork(lastProof int64) int64 {
	var proof int64 = 0
	for !bc.ValidProof(lastProof, proof) {
		proof += 1
	}
	return proof
}

func (bc *Blockchain) ValidProof(lastProof, proof int64) bool {
	guess := fmt.Sprintf("%d%d", lastProof, proof)
	guessHash := ComputeHashSha256([]byte(guess))
	return guessHash[:4] == "0000"
}

func (bc *Blockchain) ValidChain(chain *[]Block) bool {
	lastBlock := (*chain)[0]
	currentIndex := 1
	for currentIndex < len(*chain) {
		block := (*chain)[currentIndex]
		if block.PreviousHash != computeHashForBlock(lastBlock) {
			return false
		}
		if !bc.ValidProof(lastBlock.Proof, block.Proof) {
			return false
		}
		lastBlock = block
		currentIndex += 1
	}
	return true
}

func (bc *Blockchain) RegisterNode(address string) bool {
	u, err := url.Parse(address)
	if err != nil {
		return false
	}
	return bc.nodes.Add(u.Host)
}

func (bc *Blockchain) ResolveConflicts() bool {
	neighbours := bc.nodes
	newChain := make([]Block, 0)

	maxLength := len(bc.chain)

	for _, node := range neighbours.Keys() {
		otherBlockchain, err := findExternalChain(node)
		if err != nil {
			continue
		}

		if otherBlockchain.Length > maxLength && bc.ValidChain(&otherBlockchain.Chain) {
			maxLength = otherBlockchain.Length
			newChain = otherBlockchain.Chain
		}
	}
	if len(newChain) > 0 {
		bc.chain = newChain
		return true
	}

	return false
}

func NewBlockchain() *Blockchain {
	newBlockchain := &Blockchain{
		chain:        make([]Block, 0),
		transactions: make([]Transaction, 0),
		nodes:        NewStringSet(),
	}
	newBlockchain.NewBlock(100, "1")
	return newBlockchain
}

func computeHashForBlock(block Block) string {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, block)
	return ComputeHashSha256(buf.Bytes())
}

type blockchainInfo struct {
	Length int     `json:"length"`
	Chain  []Block `json:"chain"`
}

func findExternalChain(address string) (blockchainInfo, error) {
	response, err := http.Get(fmt.Sprintf("http://%s/chain", address))
	if err == nil && response.StatusCode == http.StatusOK {
		var bi blockchainInfo
		if err := json.NewDecoder(response.Body).Decode(&bi); err != nil {
			return blockchainInfo{}, err
		}
		return bi, nil
	}
	return blockchainInfo{}, err
}
