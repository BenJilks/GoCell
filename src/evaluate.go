package main

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

func (table *Table) EvaluateCellReferance(expression *Expression) (float64, error) {
    table.EnsureEvaluated(expression.row, expression.column)
    cell := table.CellAt(expression.row, expression.column)

    switch cell.kind {
    case CellText:
        return -1, errors.New("Cannot operate on text")
    case CellNumber:
        return cell.number, nil
    case CellExpression:
        return cell.number, nil
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

func (table *Table) EvaluateFunction(expression *Expression) (float64, error) {
    type Argument struct {
        is_range bool
        value float64
        row int
        column int
        end_row int
        end_column int
    }

    arguments := make([]Argument, len(expression.arguments))
    for i, argument := range expression.arguments {
        if argument.kind == ExpressionRange {
            arguments[i] = Argument {
                is_range: true,
                row: argument.row,
                column: argument.column,
                end_row: argument.end_row,
                end_column: argument.end_column,
            }
        } else {
            value, err := table.EvaluateExpression(&argument)
            if err != nil {
                return -1, err
            }

            arguments[i] = Argument { value: value }
        }
    }

    switch strings.ToLower(expression.function) {
    case "sqrt":
        if len(arguments) != 1 {
            return -1, fmt.Errorf(
                "Function 'sqrt' takes 1 argument, got %d",
                len(arguments))
        }
        if arguments[0].is_range {
            return -1, errors.New(
                "Function 'sqrt' takes a value, not range")
        }

        return math.Sqrt(arguments[0].value), nil
    case "sum":
        if len(arguments) != 1 {
            return -1, fmt.Errorf(
                "Function 'sum' takes 1 argument, got %d",
                len(arguments))
        }
        if !arguments[0].is_range {
            return -1, errors.New(
                "Function 'sum' takes a range, not value")
        }

        total := 0.0
        for row := arguments[0].row; row <= arguments[0].end_row; row++ {
            for column := arguments[0].column; column <= arguments[0].end_column; column++ {
                table.EnsureEvaluated(row, column)
                cell := table.CellAt(row, column)
                total += cell.number
            }
        }

        return total, nil
    default:
        return -1, fmt.Errorf(
            "Uknown function '%s'",
            expression.function)
    }
}

func (table *Table) EvaluateExpression(expression *Expression) (float64, error) {
    switch expression.kind {
    case ExpressionAdd:
        return table.EvaluateOperation(add_operation, expression)
    case ExpressionNumber:
        return expression.number, nil
    case ExpressionCell:
        return table.EvaluateCellReferance(expression)
    case ExpressionRange:
        return -1, errors.New("Ranges can only be used in functions")
    case ExpressionFunction:
        return table.EvaluateFunction(expression)
    default:
        panic(0)
    }
}

func (table *Table) CloneCell(row int, column int, 
                              direction Direction) Cell {
    clone_row, clone_column := direction.Offset(row, column)
    table.EnsureEvaluated(clone_row, clone_column)

    clone_cell := *table.CellAt(clone_row, clone_column)
    if clone_cell.kind == CellExpression {
        clone_cell.expression.Shift(direction.Reverse())
        clone_cell.evaluationState = EvaluationPending
        clone_cell.number = 0
    }

    return clone_cell
}

func (table *Table) EnsureEvaluated(row int, column int) {
    cell := table.CellAt(row, column)
    if cell.evaluationState == EvaluationDone {
        return
    }

    if cell.evaluationState == EvaluationInProgress {
        *cell = Cell { kind: CellError, err: errors.New("Loop!") }
        return
    }

    cell.evaluationState = EvaluationInProgress
    switch cell.kind {
    case CellExpression:
        value, err := table.EvaluateExpression(&cell.expression)
        cell.number = value
        if err != nil {
            *cell = Cell { kind: CellError, err: err }
        }
    case CellClone:
        *cell = table.CloneCell(row, column, cell.direction)
        table.EnsureEvaluated(row, column)
    }

    cell.evaluationState = EvaluationDone
}

func (table *Table) Evaluate() {
    for row := 0; row < table.rows; row++ {
        for column := 0; column < table.columns; column++ {
            table.EnsureEvaluated(row, column)
        }
    }
}

