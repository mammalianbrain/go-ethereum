package main

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/ethdb-go"
	"github.com/ethereum/ethutil-go"
	"os"
	"strings"
)

type Console struct {
	db   *ethdb.MemDatabase
	trie *ethutil.Trie
}

func NewConsole() *Console {
	db, _ := ethdb.NewMemDatabase()
	trie := ethutil.NewTrie(db, "")

	return &Console{db: db, trie: trie}
}

func (i *Console) ValidateInput(action string, argumentLength int) error {
	err := false
	var expArgCount int

	switch {
	case action == "update" && argumentLength != 2:
		err = true
		expArgCount = 2
	case action == "get" && argumentLength != 1:
		err = true
		expArgCount = 1
	case action == "dag" && argumentLength != 2:
		err = true
		expArgCount = 2
	case action == "decode" && argumentLength != 1:
		err = true
		expArgCount = 1
	case action == "encode" && argumentLength != 1:
		err = true
		expArgCount = 1
	}

	if err {
		return errors.New(fmt.Sprintf("'%s' requires %d args, got %d", action, expArgCount, argumentLength))
	} else {
		return nil
	}
}

func (i *Console) PrintRoot() {
	root := ethutil.Conv(i.trie.RootT)
	if len(root.AsBytes()) != 0 {
		fmt.Println(hex.EncodeToString(root.AsBytes()))
	} else {
		fmt.Println(i.trie.RootT)
	}
}

func (i *Console) ParseInput(input string) bool {
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(bufio.ScanWords)

	count := 0
	var tokens []string
	for scanner.Scan() {
		count++
		tokens = append(tokens, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading input:", err)
	}

	if len(tokens) == 0 {
		return true
	}

	err := i.ValidateInput(tokens[0], count-1)
	if err != nil {
		fmt.Println(err)
	} else {
		switch tokens[0] {
		case "update":
			i.trie.UpdateT(tokens[1], tokens[2])

			i.PrintRoot()
		case "get":
			fmt.Println(i.trie.GetT(tokens[1]))
		case "root":
			i.PrintRoot()
		case "rawroot":
			fmt.Println(i.trie.RootT)
		case "print":
			i.db.Print()
		case "dag":
			fmt.Println(DaggerVerify(ethutil.Big(tokens[1]), // hash
				ethutil.BigPow(2, 36),   // diff
				ethutil.Big(tokens[2]))) // nonce
		case "decode":
			d, _ := ethutil.Decode([]byte(tokens[1]), 0)
			fmt.Printf("%q\n", d)
		case "encode":
			fmt.Printf("%q\n", ethutil.Encode(tokens[1]))
		case "exit", "quit", "q":
			return false
		case "help":
			fmt.Printf("COMMANDS:\n" +
				"\033[1m= DB =\033[0m\n" +
				"update KEY VALUE - Updates/Creates a new value for the given key\n" +
				"get KEY - Retrieves the given key\n" +
				"root - Prints the hex encoded merkle root\n" +
				"rawroot - Prints the raw merkle root\n" +
				"\033[1m= Dagger =\033[0m\n" +
				"dag HASH NONCE - Verifies a nonce with the given hash with dagger\n" +
				"\033[1m= Enroding =\033[0m\n" +
				"decode STR\n" +
				"encode STR\n")

		default:
			fmt.Println("Unknown command:", tokens[0])
		}
	}

	return true
}

func (i *Console) Start() {
	fmt.Printf("Eth Console. Type (help) for help\n")
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("eth >>> ")
		str, _, err := reader.ReadLine()
		if err != nil {
			fmt.Println("Error reading input", err)
		} else {
			if !i.ParseInput(string(str)) {
				return
			}
		}
	}
}
