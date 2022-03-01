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
    arguments := make([]float64, len(expression.arguments))
    for i, argument := range expression.arguments {
        var err error

        arguments[i], err = table.EvaluateExpression(&argument)
        if err != nil {
            return -1, err
        }
    }

    switch strings.ToLower(expression.function) {
    case "sqrt":
        if len(arguments) != 1 {
            return -1, fmt.Errorf(
                "Function 'sqrt' takes 1 argument, got %d",
                len(arguments))
        }

        return math.Sqrt(arguments[0]), nil
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

