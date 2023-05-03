package main

import (
    "crypto/sha256"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "time"
)

type Block struct {
    Index         int
    Transactions  []string
    Timestamp     int64
    Previous_hash string
    Nonce         int
}

func (b *Block) compute_hash() string {
    block_string, _ := json.Marshal(b)
    return fmt.Sprintf("%x", sha256.Sum256(block_string))
}

type Blockchain struct {
    Difficulty               int
    Unconfirmed_transactions []string
    Chain                    []*Block
}

func NewBlockchain() *Blockchain {
    b := &Blockchain{
        Difficulty: 2,
        Chain:      []*Block{},
    }
    b.create_genesis_block()
    return b
}

func (bc *Blockchain) create_genesis_block() {
    genesis_block := &Block{
        Index:         0,
        Transactions:  []string{},
        Timestamp:     time.Now().Unix(),
        Previous_hash: "0",
        Nonce:         0,
    }
    genesis_block.Previous_hash = genesis_block.compute_hash()
    bc.Chain = append(bc.Chain, genesis_block)
}

func (bc *Blockchain) last_block() *Block {
    return bc.Chain[len(bc.Chain)-1]
}

func (bc *Blockchain) add_block(block *Block, proof int) bool {
    if bc.valid_proof(block, proof) {
        block.Previous_hash = bc.last_block().compute_hash()
        block.Nonce = proof
        bc.Chain = append(bc.Chain, block)
        return true
    }
    return false
}

func (bc *Blockchain) valid_proof(block *Block, proof int) bool {
    block.Nonce = proof
    guess := block.compute_hash()
    return guess[:bc.Difficulty] == "00"
}

func (bc *Blockchain) proof_of_work(block *Block) int {
    proof := 0
    for !bc.valid_proof(block, proof) {
        proof++
    }
    return proof
}

func (bc *Blockchain) new_transaction(transaction string) int {
    bc.Unconfirmed_transactions = append(bc.Unconfirmed_transactions, transaction)
    return bc.last_block().Index + 1
}

func (bc *Blockchain) mine() bool {
    if len(bc.Unconfirmed_transactions) == 0 {
        return false
    }
    last_block := bc.last_block()
    new_block := &Block{
        Index:         last_block.Index + 1,
        Transactions:  bc.Unconfirmed_transactions,
        Timestamp:     time.Now().Unix(),
        Previous_hash: last_block.compute_hash(),
        Nonce:         0,
    }
    proof := bc.proof_of_work(new_block)
    bc.add_block(new_block, proof)
    bc.Unconfirmed_transactions = []string{}
    return true
}

func main() {
    bc := NewBlockchain()

    http.HandleFunc("/transaction/new", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" {
            transaction := r.FormValue("transaction")
            index := bc.new_transaction(transaction)
            w.Write([]byte(fmt.Sprintf("Transaction will be added to Block %d", index)))
        }
    })

    http.HandleFunc("/mine", func(w http.ResponseWriter, r *http.Request) {
        if bc.mine() {
            w.Write([]byte("New block mined"))
        } else {
            w.Write([]byte("No transactions to mine"))
        }
    })

    http.HandleFunc("/chain", func(w http.ResponseWriter, r *http.Request) {
        chain, _ := json.Marshal(bc.Chain)
        w.Write(chain)
    })

    http.HandleFunc("/proof", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" {
            var block Block
            decoder := json.NewDecoder(r.Body)
            decoder.Decode(&block)
            proof := bc.proof_of_work(&block)
            w.Write([]byte(strconv.Itoa(proof)))
        }
    })

    http.ListenAndServe(":8080", nil)
}
