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

func (table *Table) EvaluateExpression(expression *Expression) (float64, error) {
    switch expression.kind {
    case ExpressionAdd:
        lhs, err := table.EvaluateExpression(expression.lhs)
        if err != nil {
            return -1, err
        }

        rhs, err := table.EvaluateExpression(expression.rhs)
        if err != nil {
            return -1, err
        }
        
        return lhs + rhs, nil
    case ExpressionNumber:
        return expression.number, nil
    case ExpressionCell:
        index := expression.row * table.columns + expression.column
        cell := &table.content[index]
        switch cell.kind {
        case CellText:
            return -1, errors.New("Cannot operate on text")
        case CellNumber:
            return cell.number, nil
        case CellExpression:
            value, err := table.EvaluateCell(cell)
            if err != nil {
                return -1, errors.New(fmt.Sprintf("Error in %c%d",
                    expression.column + 'A', expression.row))
            }

            return value, nil
        case CellError:
            return -1, cell.err
        case CellEmpty:
            return 0, nil
        default:
            panic(0)
        }
    default:
        panic(0)
    }
}

func (table *Table) EvaluateCell(cell *Cell) (float64, error) {
    if cell.is_evaluating {
        err := errors.New("Loop!")
        *cell = Cell {
            kind: CellError,
            err: err,
        }
        return -1, err
    }

    cell.is_evaluating = true
    defer func() { cell.is_evaluating = false }()

    value, err := table.EvaluateExpression(&cell.expression)
    if err != nil {
        *cell = Cell {
            kind: CellError,
            err: err,
        }
    } else {
        *cell = Cell {
            kind: CellNumber,
            number: value,
        }
    }

    return value, err
}

func (table *Table) Evaluate() {
    for row := 0; row < table.rows; row++ {
        for column := 0; column < table.columns; column++ {
            index := row * table.columns + column
            cell := &table.content[index]
            if cell.kind != CellExpression {
                continue
            }

            table.EvaluateCell(cell)
        }
    }
}

func (table *Table) Print() {
    widths := make([]int, table.columns)
    for row := 0; row < table.rows; row++ {
        for column := 0; column < table.columns; column++ {
            index := row * table.columns + column
            cell := table.content[index]
            width := len(cell.String())
            if width > widths[column] {
                widths[column] = width
            }
        }
    }

    for row := 0; row < table.rows; row++ {
        for column := 0; column < table.columns; column++ {
            index := row * table.columns + column
            cell := table.content[index]
            text := cell.String()

            fmt.Printf("%-" + fmt.Sprint(widths[column]) + "s", text)
            if column != table.columns - 1 {
                fmt.Print(" | ")
            }
        }
        fmt.Println()
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

func readTable(filePath string) Table {
    input_bytes, err := os.ReadFile(filePath)
    if err != nil {
        panic(err)
    }
    
    input := string(input_bytes)
    rows, columns := countTableSize(input)
    content := readTableContent(input, rows, columns)

    return Table {
        content,
        rows,
        columns,
    }
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Error: Expected input file")
        os.Exit(1)
    }

    inputFile := os.Args[1]
    table := readTable(inputFile)
    table.Evaluate()
    table.Print()
}

