package main

import (
	"fmt"
	"os"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Error: Expected input file")
        os.Exit(1)
    }

    inputFile := os.Args[1]
    table, err := readTable(inputFile)
    if err != nil {
        fmt.Println("Error:", err)
        os.Exit(1)
    }

    table.Evaluate()
    table.Print()
}

