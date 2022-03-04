package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
)

type Table struct {
    allocator ExpressionAllocator
    content []Cell
    rows int
    columns int
}

func (table *Table) CellAt(position CellPosition) *Cell {
    if position.row < 0 || position.row >= table.rows ||
        position.column < 0 || position.column >= table.columns {
        return &Cell {
            kind: CellError,
            err: errors.New("Cell outside table"),
        }
    }

    index := position.row * table.columns + position.column
    return &table.content[index]
}

func (table *Table) IsEmpty(position CellPosition) bool {
    if position.row < 0 || position.row >= table.rows ||
        position.column < 0 || position.column >= table.columns {
        return false
    }

    index := position.row * table.columns + position.column
    return table.content[index].kind == CellEmpty
}

func cloneCell(allocator *ExpressionAllocator,
               cell Cell,
               direction Direction,
               offset int) Cell {
    if cell.kind == CellExpression {
        cell.expressionOffset = cell.expressionOffset.Offset(
            direction.Reverse(), offset)
        cell.evaluationState = EvaluationPending
        cell.number = 0
    }

    return cell
}

func (table *Table) Print(output io.Writer) {
    cellTexts := make([]string, table.rows * table.columns)
    widths := make([]int, table.columns)
    for row := 0; row < table.rows; row++ {
        for column := 0; column < table.columns; column++ {
            index := row * table.columns + column
            cell := table.content[index].String()
            cellTexts[index] = cell

            width := len(cell)
            if width > widths[column] {
                widths[column] = width
            }
        }
    }

    for row := 0; row < table.rows; row++ {
        last_non_empty := table.columns - 1
        for last_non_empty > 0 && table.IsEmpty(CellPosition { row, last_non_empty - 1 }) {
            last_non_empty -= 1
        }

        for column := 0; column < last_non_empty + 1; column++ {
            index := row * table.columns + column
            text, cell := cellTexts[index], table.content[index]
            if cell.kind == CellSeporator {
                text = strings.Repeat("_", widths[column])
            }

            padding := widths[column] - len(text)
            output.Write(bytes.Repeat([]byte{ ' ' }, padding))
            output.Write([]byte(text))
            if column != last_non_empty {
                output.Write([]byte(" | "))
            }
        }
        output.Write([]byte{ '\n' })
    }
}

func countTableSize(input string) (int, int) {
    rowCount := 0
    maxColumnCount := 0

    scanner := bufio.NewScanner(strings.NewReader(input))
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if len(line) >= 3 && line[:3] == "..." {
            count, _ := strconv.ParseUint(strings.TrimSpace(line[3:]), 10, 32)
            rowCount += int(count)
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

func readCloneLine(allocator *ExpressionAllocator,
                   content []Cell,
                   columns int,
                   line string,
                   row int) (int, error) {
    count, err := strconv.ParseUint(strings.TrimSpace(line[3:]), 10, 32)
    if err != nil {
        return row, err
    }
    if row == 0 {
        return row, errors.New("Cannot duplicate above the table")
    }

    for i := 0; i < int(count); i++ {
        for column := 0; column < columns; column++ {
            above_index := (row - 1) * columns + column
            index := row * columns + column
            content[index] = cloneCell(allocator, content[above_index], DirectionUp, 1)
        }
        row += 1
    }

    return row, nil
}

func readTableContent(allocator *ExpressionAllocator,
                      input string,
                      rows int,
                      columns int) ([]Cell, error) {
    content := make([]Cell, rows * columns)
    for i := range content {
        content[i] = Cell { kind: CellEmpty }
    }

    scanner := bufio.NewScanner(strings.NewReader(input))
    row := 0
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if len(line) >= 3 && line[:3] == "..." {
            new_row, err := readCloneLine(allocator, content, columns, line, row)
            if err != nil {
                return nil, err
            }

            row = new_row
            continue
        }

        cells := strings.Split(line, "|")
        for column := 0; column < len(cells); column++ {
            text := strings.TrimSpace(cells[column])
            cell := parseCell(allocator, text, CellPosition { row, column })
            index := row * columns + column
            content[index] = cell
        }

        row += 1
    }

    return content, nil
}

func readTable(filePath string) (Table, error) {
    input_bytes, err := os.ReadFile(filePath)
    if err != nil {
        return Table{}, err
    }
    
    input := string(input_bytes)
    rows, columns := countTableSize(input)
    allocator := newExpressionAllocator()
    content, err := readTableContent(&allocator, input, rows, columns)
    if err != nil {
        return Table{}, err
    }

    return Table {
        allocator,
        content,
        rows,
        columns,
    }, nil
}

