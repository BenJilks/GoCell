package main

import (
	"errors"
	"fmt"
	"strconv"
)

type Direction int
const (
    DirectionUp Direction = iota
    DirectionRight
    DirectionDown
    DirectionLeft
    DirectionNone
)

func (direction Direction) Reverse() Direction {
    switch direction {
    case DirectionUp: return DirectionDown
    case DirectionRight: return DirectionLeft
    case DirectionDown: return DirectionUp
    case DirectionLeft: return DirectionRight
    default: panic(0)
    }
}

func (direction Direction) IsUpOrLeft() bool {
    switch direction {
    case DirectionUp, DirectionLeft: return true
    case DirectionDown, DirectionRight: return false
    default: panic(0)
    }
}

type CellKind int
const (
    CellText CellKind = iota
    CellNumber
    CellExpression
    CellClone
    CellSeporator
    CellError
    CellEmpty
)

type EvaluationState int
const (
    EvaluationPending EvaluationState = iota
    EvaluationInProgress
    EvaluationDone
)

type Cell struct {
    kind CellKind
    evaluationState EvaluationState

    text string
    number float64
    direction Direction
    offset int
    err error

    expression *Expression
    expressionOffset CellPosition
}

type CellPosition struct {
    row int
    column int
}

func (position CellPosition) Offset(direction Direction, count int) CellPosition {
    row, column := position.row, position.column
    switch direction {
    case DirectionUp:
        return CellPosition { row - count, column }
    case DirectionRight:
        return CellPosition { row, column + count }
    case DirectionDown:
        return CellPosition { row + count, column }
    case DirectionLeft:
        return CellPosition { row, column - count }
    default:
        panic(0)
    }
}

func formatCellNumber(number float64) string {
    return strconv.FormatFloat(number, 'f', -1, 32)
}

func (cell Cell) String() string {
    switch cell.kind {
    case CellText:
        return cell.text
    case CellNumber:
        return formatCellNumber(cell.number)
    case CellExpression:
        if cell.evaluationState == EvaluationDone {
            return formatCellNumber(cell.number)
        } else {
            return "#ERROR#"
        }
    case CellSeporator:
        return ""
    case CellError:
        return "#" + fmt.Sprint(cell.err) + "#"
    case CellEmpty:
        return ""
    default:
        panic(0)
    }
}

func (cell *Cell) Offset(direction Direction, offset int) {
    if cell.kind == CellExpression {
        cell.expressionOffset = cell.expressionOffset.Offset(
            direction.Reverse(), offset)
        cell.evaluationState = EvaluationPending
        cell.number = 0
    }
}

func parseDirection(text string) (Direction, error) {
    if len(text) == 0 {
        return DirectionNone, errors.New("No clone direction")
    }

    switch text[0] {
        case '^':
            return DirectionUp, nil
        case '>':
            return DirectionRight, nil
        case 'v':
            return DirectionDown, nil
        case '<':
            return DirectionLeft, nil
        default:
            return DirectionNone, fmt.Errorf(
                "Unvalid clone direction '%s'", text)
    }
}

func parseCell(allocator *ExpressionAllocator,
               text string,
               position CellPosition) Cell {
    if len(text) == 0 {
        return Cell { kind: CellEmpty }
    }

    if text[0] == '=' {
        expression, _, err := parseExpression(allocator, text[1:], position)
        if err != nil {
            return Cell { kind: CellError, err: err }
        }

        return Cell { kind: CellExpression, expression: expression }
    }

    if text[0] == ':' {
        direction, err := parseDirection(text[1:]) 
        if err != nil {
            return Cell { kind: CellError, err: err }
        }

        offset := 1
        if len(text) > 2 && isDigit(text[2]) {
            offset, _, err = parseInt(text[2:])
        }
        return Cell { kind: CellClone, direction: direction, offset: offset }
    }

    if len(text) > 1 && text[:2] == "__" {
        return Cell { kind: CellSeporator }
    }

    number, err := strconv.ParseFloat(text, 64)
    if err == nil {
        return Cell { kind: CellNumber, number: number }
    }

    return Cell { kind: CellText, text: text }
}

