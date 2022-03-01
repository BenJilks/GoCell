package main

import (
	"fmt"
	"strconv"
)

type CellKind int
const (
    CellText CellKind = iota
    CellNumber
    CellExpression
    CellError
    CellEmpty
)

type Cell struct {
    kind CellKind
    is_evaluating bool

    text string
    number float64
    expression Expression
    err error
}

func (cell Cell) String() string {
    switch cell.kind {
    case CellText:
        return cell.text
    case CellNumber:
        return fmt.Sprintf("%f", cell.number)
    case CellExpression:
        return "#ERROR#"
    case CellError:
        return "#" + fmt.Sprint(cell.err) + "#"
    case CellEmpty:
        return ""
    default:
        panic(0)
    }
}

func parseCell(text string) Cell {
    if len(text) == 0 {
        return Cell { kind: CellEmpty }
    }

    if text[0] == '=' {
        expression, _, err := parseExpression(text[1:]) 
        if err != nil {
            return Cell {
                kind: CellError,
                err: err,
            }
        }

        return Cell {
            kind: CellExpression,
            expression: expression,
        }
    }

    number, err := strconv.ParseFloat(text, 64)
    if err == nil {
        return Cell {
            kind: CellNumber,
            number: number,
        }
    }

    return Cell {
        kind: CellText,
        text: text,
    }
}

