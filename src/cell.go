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
    case DirectionUp:
        return DirectionDown
    case DirectionRight:
        return DirectionLeft
    case DirectionDown:
        return DirectionUp
    case DirectionLeft:
        return DirectionRight
    default:
        panic(0)
    }
}

func (direction Direction) Offset(row int, column int) (int, int) {
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

type CellKind int
const (
    CellText CellKind = iota
    CellNumber
    CellExpression
    CellClone
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
    expression Expression
    direction Direction
    err error
}

func (cell Cell) String() string {
    switch cell.kind {
    case CellText:
        return cell.text
    case CellNumber:
        return fmt.Sprintf("%.6g", cell.number)
    case CellExpression:
        if cell.evaluationState == EvaluationDone {
            return fmt.Sprintf("%.6g", cell.number)
        } else {
            return "#ERROR#"
        }
    case CellError:
        return "#" + fmt.Sprint(cell.err) + "#"
    case CellEmpty:
        return ""
    default:
        panic(0)
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

func parseCell(text string) Cell {
    if len(text) == 0 {
        return Cell { kind: CellEmpty }
    }

    if text[0] == '=' {
        expression, _, err := parseExpression(text[1:]) 
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

        return Cell { kind: CellClone, direction: direction }
    }

    number, err := strconv.ParseFloat(text, 64)
    if err == nil {
        return Cell { kind: CellNumber, number: number }
    }

    return Cell { kind: CellText, text: text }
}

