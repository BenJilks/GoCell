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

func (table *Table) EvaluateCellReferance(expression *Expression) (float64, error) {
    index := expression.row * table.columns + expression.column
    cell := &table.content[index]

    switch cell.kind {
    case CellText:
        return -1, errors.New("Cannot operate on text")
    case CellNumber:
        return cell.number, nil
    case CellExpression:
        value, err := table.EvaluateCell(expression.row, expression.column)
        if err != nil {
            return -1, fmt.Errorf(
                "Error in %c%d",
                expression.column + 'A',
                expression.row)
        }

        return value, nil
    case CellError:
        return -1, cell.err
    case CellEmpty:
        return 0, nil
    default:
        panic(0)
    }
}

func (table *Table) EvaluateOperation(operation func(float64, float64) float64,
                                      expression *Expression) (float64, error) {
    lhs, err := table.EvaluateExpression(expression.lhs)
    if err != nil {
        return -1, err
    }

    rhs, err := table.EvaluateExpression(expression.rhs)
    if err != nil {
        return -1, err
    }
    
    return operation(lhs, rhs), nil
}

func add_operation(a float64, b float64) float64 {
    return a + b
}

func (table *Table) EvaluateExpression(expression *Expression) (float64, error) {
    switch expression.kind {
    case ExpressionAdd:
        return table.EvaluateOperation(add_operation, expression)
    case ExpressionNumber:
        return expression.number, nil
    case ExpressionCell:
        return table.EvaluateCellReferance(expression)
    default:
        panic(0)
    }
}

func cellInDirection(row int, column int,
                     direction Direction) (int, int) {
    switch direction {
    case DirectionUp:
        return row - 1, column
    case DirectionRight:
        return row, column + 1
    case DirectionDown:
        return row + 1, column
    case DirectionLeft:
        return row, column - 1
    default:
        panic(0)
    }
}

func (table *Table) cloneCell(row int, column int, 
                              direction Direction) Cell {
    clone_row, clone_column := cellInDirection(row, column, direction)
    if clone_row < 0 || clone_row >= table.rows ||
        clone_column < 0 || clone_column >= table.columns {
        return Cell {
            kind: CellError,
            err: errors.New("Clone outside table"),
        }
    }

    table.EvaluateCell(clone_row, clone_column)
    index := clone_row * table.columns + clone_column
    clone_cell := table.content[index]
    if clone_cell.kind == CellExpression {
        clone_cell.expression.shift(direction.reverse())
        clone_cell.evaluationState = EvaluationPending
        clone_cell.number = 0
    }

    return clone_cell
}

func (table *Table) EvaluateCell(row int, column int) (float64, error) {
    cell := &table.content[row * table.columns + column]
    if cell.evaluationState == EvaluationDone {
        return cell.number, nil
    }

    if cell.evaluationState == EvaluationInProgress {
        *cell = Cell { kind: CellError, err: errors.New("Loop!") }
        return -1, cell.err
    }

    cell.evaluationState = EvaluationInProgress
    switch cell.kind {
    case CellExpression:
        value, err := table.EvaluateExpression(&cell.expression)
        cell.number = value
        cell.evaluationState = EvaluationDone
        if err != nil {
            *cell = Cell { kind: CellError, err: err }
        }

        return value, err
    case CellClone:
        *cell = table.cloneCell(row, column, cell.direction)
        return table.EvaluateCell(row, column)
    default:
        return 0, nil
    }
}

func (table *Table) Evaluate() {
    for row := 0; row < table.rows; row++ {
        for column := 0; column < table.columns; column++ {
            table.EvaluateCell(row, column)
        }
    }
}

func (table *Table) Print() {
    content := make([][]string, table.rows)
    widths := make([]int, table.columns)
    for row := 0; row < table.rows; row++ {
        content[row] = make([]string, table.columns)
        for column := 0; column < table.columns; column++ {
            index := row * table.columns + column
            cell := table.content[index]
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

