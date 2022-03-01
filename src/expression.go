package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type TokenKind int
const (
    TokenAdd TokenKind = iota
    TokenNumber
    TokenCell
    TokenEmpty
)

type Token struct {
    kind TokenKind
    number float64
    row int
    column int
}

type ExpressionKind int
const (
    ExpressionAdd ExpressionKind = iota
    ExpressionNumber
    ExpressionCell
)

type Expression struct {
    kind ExpressionKind
    lhs *Expression
    rhs *Expression
    row int
    column int
    number float64
}

func (expression *Expression) Shift(direction Direction) {
    switch expression.kind {
    case ExpressionAdd:
        expression.lhs.Shift(direction)
        expression.rhs.Shift(direction)
    case ExpressionCell:
        row, column := &expression.row, &expression.column
        *row, *column = direction.Offset(*row, *column)
    case ExpressionNumber:
        break
    }
}

func parseCellReference(text string) (Token, string, error) {
    column := text[0] - 'A'

    i := 1
    for i < len(text) && text[i] != ' ' {
        i += 1
    }

    row, err := strconv.ParseInt(text[1:i], 10, 32)
    if err != nil {
        return Token{}, text, err
    }

    return Token {
        kind: TokenCell,
        row: int(row) - 1,
        column: int(column),
    }, text[i:], nil
}

func parseNumber(text string) (Token, string, error) {
    i := 1
    for i < len(text) && text[i] != ' ' {
        i += 1
    }

    number, err := strconv.ParseFloat(text[:i], 64)
    if err != nil {
        return Token{}, text, err
    }

    return Token {
        kind: TokenNumber,
        number: number,
    }, text[i:], nil
}

func nextToken(text string) (Token, string, error) {
    text = strings.TrimLeft(text, " ")
    if len(text) == 0 {
        return Token { kind: TokenEmpty }, text, nil
    }

    switch c := text[0]
    {
    case c == '+':
        return Token { kind: TokenAdd }, text[1:], nil
    case c >= 'A' && c <= 'Z':
        return parseCellReference(text)
    case c >= '0' && c <= '9':
        return parseNumber(text)
    default:
        return Token {}, text[1:], fmt.Errorf(
            "Unexpected char '%c'", c)
    }
}

func parseTerm(text string) (Expression, string, error) {
    token, text, err := nextToken(text)
    if err != nil {
        return Expression{}, text, err
    }

    switch token.kind {
    case TokenNumber:
        return Expression {
            kind: ExpressionNumber,
            number: token.number,
        }, text, nil
    case TokenCell:
        return Expression {
            kind: ExpressionCell,
            row: token.row,
            column: token.column,
        }, text, nil
    case TokenAdd:
        return Expression{}, text, errors.New(
            "Unexpected '+', expected value")
    case TokenEmpty:
        return Expression{}, text, errors.New(
            "Expected value, got nothing instead")
    default:
        panic(0)
    }
}

func parseOperation(lhs Expression,
                    kind ExpressionKind,
                    text string) (Expression, string, error) {
    rhs, text, err := parseExpression(text)
    if err != nil {
        return Expression{}, text, err
    }

    return Expression {
        kind: kind,
        lhs: &lhs,
        rhs: &rhs,
    }, text, nil
}

func parseExpression(text string) (Expression, string, error) {
    result, text, err := parseTerm(text)
    if err != nil {
        return Expression{}, text, err
    }

    var token Token
    for token.kind != TokenEmpty {
        token, text, err = nextToken(text)
        if err != nil {
            return Expression{}, text, err
        }

        switch token.kind {
        case TokenEmpty:
            break
        case TokenAdd:
            result, text, err = parseOperation(result, ExpressionAdd, text)
            if err != nil {
                return Expression{}, text, err
            }
        case TokenNumber:
            return Expression{}, text, fmt.Errorf(
                "Unexpected '%.6g', expected operator",
                token.number)
        case TokenCell:
            return Expression{}, text, fmt.Errorf(
                "Unexpected '%c%d', expected operator",
                token.column + 'A',
                token.row + 1)
        }
    }

    return result, text, nil
}

