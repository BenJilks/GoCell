package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Table struct {
    content []Cell
    rows int
    columns int
}

func (table *Table) CellAt(row int, column int) *Cell {
    if row < 0 || row >= table.rows ||
        column < 0 || column >= table.columns {
        return &Cell {
            kind: CellError,
            err: errors.New("Cell outside table"),
        }
    }

    index := row * table.columns + column
    return &table.content[index]
}

func (table *Table) Print() {
    content := make([][]string, table.rows)
    widths := make([]int, table.columns)
    for row := 0; row < table.rows; row++ {
        content[row] = make([]string, table.columns)
        for column := 0; column < table.columns; column++ {
            cell := table.CellAt(row, column)
            text := cell.String()
            content[row][column] = text

            width := len(text)
            if width > widths[column] {
                widths[column] = width
            }
        }
    }

    for _, row := range content {
        padded_row := make([]string, table.columns)
        for column, cell := range row {
            format := fmt.Sprintf("%%-%ds", column)
            padded_row[column] = fmt.Sprintf(format, cell)
        }

        fmt.Println(strings.Join(padded_row, " | "))
    }
}

func countTableSize(input string) (int, int) {
    rowCount := 0
    maxColumnCount := 0

    scanner := bufio.NewScanner(strings.NewReader(input))
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if len(line) == 0 {
            continue
        }

        columnCount := strings.Count(line, "|") + 1
        rowCount += 1
        if columnCount > maxColumnCount {
            maxColumnCount = columnCount
        }
    }

    return rowCount, maxColumnCount
}

func readTableContent(input string, rows int, columns int) []Cell {
    content := make([]Cell, rows * columns)
    for i := range content {
        content[i] = Cell { kind: CellEmpty }
    }

    scanner := bufio.NewScanner(strings.NewReader(input))
    row := 0
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if len(line) == 0 {
            continue
        }

        cells := strings.Split(line, "|")
        for column := 0; column < len(cells); column++ {
            text := strings.TrimSpace(cells[column])
            cell := parseCell(text)
            index := row * columns + column
            content[index] = cell
        }

        row += 1
    }

    return content
}

func readTable(filePath string) (Table, error) {
    input_bytes, err := os.ReadFile(filePath)
    if err != nil {
        return Table{}, err
    }
    
    input := string(input_bytes)
    rows, columns := countTableSize(input)
    content := readTableContent(input, rows, columns)
    return Table {
        content,
        rows,
        columns,
    }, nil
}

