package main

import (
	"flag"
	"fmt"
	"github.com/nd-sin/blockchain/blockchain"
	"log"
	"os"
	"runtime"
	"strconv"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("init -address ADDRESS - initialize database with genesis block")
	fmt.Println("blockchain -address ADDRESS - creates a blockchain")
	fmt.Println("print - Prints the blocks in the chain")
	fmt.Println("print - Prints the blocks in the chain")
	fmt.Println("send -from FROM - to TO -amount AMOUNT - Send amount")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) printChain() {
	chain := blockchain.ContinueBlockchain("")
	defer chain.Database.Close()
	iter := chain.Iterator()
	for {
		block := iter.Next()
		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Current Hash: %x\n", block.Hash)
		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println("===========")
		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) createBlockchain(address string) {
	chain := blockchain.ContinueBlockchain(address)
	chain.Database.Close()
	fmt.Println("Finished!")
}

func (cli *CommandLine) initBlockchain(address string) {
	chain := blockchain.InitBlockchain(address)
	chain.Database.Close()
	fmt.Println("Finished!")
}

func (cli *CommandLine) send(from, to string, amount int) {
	chain := blockchain.ContinueBlockchain(from)
	defer chain.Database.Close()
	tx := blockchain.NewTransaction(from, to, amount, chain)
	chain.AddBlock([]*blockchain.Transaction{tx})
	fmt.Println("Success!")
}

func (cli *CommandLine) getBalance(address string) {
	chain := blockchain.ContinueBlockchain(address)
	defer chain.Database.Close()
	balance := 0
	UTXOs := chain.FindUTXO(address)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) run() {
	cli.validateArgs()
	getBalanceCmd := flag.NewFlagSet("balance", flag.ExitOnError)
	initBlockchainCmd := flag.NewFlagSet("init", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("blockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	initBlockchainAddress := initBlockchainCmd.String("address", "", "The address to create for genesis block")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	switch os.Args[1] {
	case "balance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "blockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "init":
		err := initBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		runtime.Goexit()
	}
	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress)
	}
	if initBlockchainCmd.Parsed() {
		if *initBlockchainAddress == "" {
			initBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.initBlockchain(*initBlockchainAddress)
	}
	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockchain(*createBlockchainAddress)
	}
	if printChainCmd.Parsed() {
		cli.printChain()
	}
	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}

func main() {
	defer os.Exit(0)
	cli := CommandLine{}
	cli.run()
}
