package main

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

func (table *Table) EvaluateCellReferance(expression *Expression,
                                          shiftOffset CellPosition) (float64, error) {
    position := CellPosition {
        expression.position.row + shiftOffset.row,
        expression.position.column + shiftOffset.column,
    }

    table.EnsureEvaluated(position)
    cell := table.CellAt(position)

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
                                      expression *Expression,
                                      shiftOffset CellPosition) (float64, error) {
    lhs, err := table.EvaluateExpression(expression.lhs, shiftOffset)
    if err != nil {
        return -1, err
    }

    rhs, err := table.EvaluateExpression(expression.rhs, shiftOffset)
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
    function func(*Table, []*Expression, CellPosition) float64
    expected_arguments []bool
}

func sum(table *Table,
         arguments []*Expression,
         shiftOffset CellPosition) float64 {
    total := 0.0
    a := arguments[0].cellRange
    for row := a.start.row; row <= a.end.row; row++ {
        for column := a.start.column; column <= a.end.column; column++ {
            table.EnsureEvaluated(CellPosition { row, column })
            total += table.CellAt(CellPosition { row, column }).number
        }
    }

    return total
}

func sqrt(table *Table,
          arguments []*Expression,
          shiftOffset CellPosition) float64 {
    value, _ := table.EvaluateExpression(
        arguments[0], shiftOffset)
    return math.Sqrt(value)
}

func (table *Table) GetFunction(name string) (Function, bool) {
    switch strings.ToLower(name) {
    case "sum":
        return Function {
            function: sum,
            expected_arguments: []bool { true },
        }, true
    case "sqrt":
        return Function {
            function: sqrt,
            expected_arguments: []bool { false },
        }, true
    default:
        return Function{}, false
    }
}

func validateArguments(name string, function Function, arguments []*Expression) error {
    if len(arguments) != len(function.expected_arguments) {
        return fmt.Errorf(
            "Function '%s' takes %d argument(s), got %d",
            name,
            len(function.expected_arguments),
            len(arguments))
    }

    for i := range arguments {
        if (arguments[i].kind == ExpressionRange) != function.expected_arguments[i] {
            return fmt.Errorf(
                "Function '%s' takes a value, not range",
                name)
        }
    }

    return nil
}

func (table *Table) EvaluateFunction(expression *Expression,
                                     shiftOffset CellPosition) (float64, error) {
    name := expression.function
    function, found := table.GetFunction(name)
    if !found {
        return -1, fmt.Errorf("Uknown function '%s'", name)
    }

    if err := validateArguments(name, function, expression.arguments); err != nil {
        return -1, nil
    }

    return function.function(table, expression.arguments, shiftOffset), nil
}

func (table *Table) EvaluateExpression(expression *Expression,
                                       shiftOffset CellPosition) (float64, error) {
    switch expression.kind {
    case ExpressionAdd:
        return table.EvaluateOperation(
            add_operation, expression, shiftOffset)
    case ExpressionNumber:
        return expression.number, nil
    case ExpressionCell:
        return table.EvaluateCellReferance(expression, shiftOffset)
    case ExpressionConstant:
        return table.EvaluateCellReferance(expression, CellPosition{})
    case ExpressionRange:
        return -1, errors.New("Ranges can only be used in functions")
    case ExpressionFunction:
        return table.EvaluateFunction(expression, shiftOffset)
    default:
        panic(0)
    }
}

func (table *Table) EnsureEvaluated(position CellPosition) {
    cell := table.CellAt(position)
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
        value, err := table.EvaluateExpression(
            cell.expression, cell.expressionOffset)

        cell.number = value
        if err != nil {
            *cell = Cell { kind: CellError, err: err }
        }
    case CellClone:
        clone_position := position.Offset(cell.direction, cell.offset)
        table.EnsureEvaluated(clone_position)

        clone_cell := *table.CellAt(clone_position)
        *cell = cloneCell(&table.allocator, clone_cell, cell.direction, cell.offset)
        table.EnsureEvaluated(position)
    }

    cell.evaluationState = EvaluationDone
}

func (table *Table) Evaluate() {
    for row := 0; row < table.rows; row++ {
        for column := 0; column < table.columns; column++ {
            table.EnsureEvaluated(CellPosition { row, column })
        }
    }
}

