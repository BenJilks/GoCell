package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

func main() {
    outputFile := flag.String("output", "", "Specify output file")
    help := flag.Bool("help", false, "Show help screen")
    flag.Parse()

    inputFile := flag.Arg(0)
    if inputFile == "" {
        fmt.Fprintln(os.Stderr, "No input file given")
        flag.Usage()
        os.Exit(1)
    }

    if *help {
        flag.Usage()
        os.Exit(1)
    }

    table, err := readTable(inputFile)
    if err != nil {
        fmt.Println("Error:", err)
        os.Exit(1)
    }

    table.Evaluate()

    if *outputFile == "" {
        table.Print(os.Stdout)
    } else {
        file, err := os.Create(*outputFile)
        if err != nil {
            panic(err)
        }

        writer := bufio.NewWriter(file)
        table.Print(writer)
        writer.Flush()
        file.Close()
    }
}

