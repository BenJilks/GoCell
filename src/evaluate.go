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

type Argument struct {
    is_range bool
    value float64
    cellRange Range
}

type Function struct {
    function func(*Table, []Argument) float64
    expected_arguments []bool
}

func sum(table *Table, arguments []Argument) float64 {
    total := 0.0
    a := arguments[0].cellRange
    for row := a.start_row; row <= a.end_row; row++ {
        for column := a.start_column; column <= a.end_column; column++ {
            table.EnsureEvaluated(row, column)
            total += table.CellAt(row, column).number
        }
    }

    return total
}

func sqrt(_ *Table, arguments []Argument) float64 {
    return math.Sqrt(arguments[0].value)
}

func (table *Table) GetFunction(name string) *Function {
    switch strings.ToLower(name) {
    case "sum":
        return &Function {
            function: sum,
            expected_arguments: []bool { true },
        }
    case "sqrt":
        return &Function {
            function: sqrt,
            expected_arguments: []bool { false },
        }
    default:
        return nil
    }
}

func validateArguments(name string, function *Function, arguments []Argument) error {
    if function == nil {
        return fmt.Errorf("Uknown function '%s'", name)
    }

    if len(arguments) != len(function.expected_arguments) {
        return fmt.Errorf(
            "Function '%s' takes %d argument(s), got %d",
            name,
            len(function.expected_arguments),
            len(arguments))
    }

    for i := range arguments {
        if arguments[i].is_range != function.expected_arguments[i] {
            return fmt.Errorf(
                "Function '%s' takes a value, not range",
                name)
        }
    }

    return nil
}

func (table *Table) EvaluateFunction(expression *Expression) (float64, error) {
    arguments := make([]Argument, len(expression.arguments))
    for i, argument := range expression.arguments {
        if argument.kind == ExpressionRange {
            arguments[i] = Argument {
                is_range: true,
                cellRange: argument.cellRange,
            }
            continue
        }

        value, err := table.EvaluateExpression(&argument)
        if err != nil {
            return -1, err
        }

        arguments[i] = Argument { value: value }
    }

    name := expression.function
    function := table.GetFunction(name)
    if err := validateArguments(name, function, arguments); err != nil {
        return -1, nil
    }

    return function.function(table, arguments), nil
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

