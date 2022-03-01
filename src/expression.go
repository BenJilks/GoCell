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
    TokenOpenBrace
    TokenCloseBrace
    TokenComma
    TokenName
    TokenNumber
    TokenCell
    TokenRange
    TokenEmpty
)

type Token struct {
    kind TokenKind
    name string
    number float64
    row int
    column int
    end_row int
    end_column int
}

type ExpressionKind int
const (
    ExpressionAdd ExpressionKind = iota
    ExpressionNumber
    ExpressionCell
    ExpressionRange
    ExpressionFunction
)

type Expression struct {
    kind ExpressionKind
    number float64

    lhs *Expression
    rhs *Expression

    row int
    column int
    end_row int
    end_column int

    function string
    arguments []Expression
}

func (expression *Expression) Shift(direction Direction) {
    switch expression.kind {
    case ExpressionAdd:
        expression.lhs.Shift(direction)
        expression.rhs.Shift(direction)
    case ExpressionCell:
        row, column := &expression.row, &expression.column
        *row, *column = direction.Offset(*row, *column)
    case ExpressionRange:
        row, column := &expression.row, &expression.column
        end_row, end_column := &expression.end_row, &expression.end_column
        *row, *column = direction.Offset(*row, *column)
        *end_row, *end_column = direction.Offset(*end_row, *end_column)
    case ExpressionFunction:
        for i := range expression.arguments {
            expression.arguments[i].Shift(direction)
        }
    case ExpressionNumber:
    default:
        panic(0)
    }
}

func isLetter(c byte) bool {
    return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func isDigit(c byte) bool {
    return c >= '0' && c <= '9'
}

func parseName(text string) (Token, string, error) {
    i := 0
    for i < len(text) && isLetter(text[i]) {
        i += 1
    }

    if i < len(text) && isDigit(text[i]) {
        return Token{}, text, errors.New("Not a name")
    }

    return Token {
        kind: TokenName,
        name: text[:i],
    }, text[i:], nil
}

func parseRange(row int, column int, text string) (Token, string, error) {
    end, text, err := parseCellReference(text)
    if err != nil {
        return Token{}, text, err
    }

    return Token {
        kind: TokenRange,
        row: int(row) - 1,
        column: int(column),
        end_row: end.row,
        end_column: end.column,
    }, text, nil
}

func parseCellReference(text string) (Token, string, error) {
    column_name := strings.ToUpper(text)[0]
    column := int(column_name - 'A')

    i := 1
    for i < len(text) && isDigit(text[i]) {
        i += 1
    }

    row_index, err := strconv.ParseInt(text[1:i], 10, 32)
    if err != nil {
        return Token{}, text, err
    }

    row := int(row_index) - 1
    if i < len(text) && text[i] == ':' {
        return parseRange(row, column, text[i+1:])
    }

    return Token {
        kind: TokenCell,
        row: row,
        column: column,
    }, text[i:], nil
}

func parseNumber(text string) (Token, string, error) {
    i := 0
    for i < len(text) && isDigit(text[i]) {
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
    case c == '(':
        return Token { kind: TokenOpenBrace }, text[1:], nil
    case c == ')':
        return Token { kind: TokenCloseBrace }, text[1:], nil
    case c == ',':
        return Token { kind: TokenComma }, text[1:], nil
    case isLetter(c):
        if name, text, err := parseName(text); err == nil {
            return name, text, err
        } else {
            return parseCellReference(text)
        }
    case isDigit(c):
        return parseNumber(text)
    default:
        return Token {}, text[1:], fmt.Errorf(
            "Unexpected char '%c'", c)
    }
}

func expect(kind TokenKind, text string) (string, error) {
    token, text, err := nextToken(text)
    if err != nil {
        return text, err
    }

    if token.kind != kind {
        return text, fmt.Errorf(
            "Expected '%d', got '%d' instead",
            kind, token.kind)
    }

    return text, nil
}

func parseFunction(function string, text string) (Expression, string, error) {
    text, err := expect(TokenOpenBrace, text)
    if err != nil {
        return Expression{}, text, err
    }
    
    arguments := make([]Expression, 0)
    for {
        var argument Expression
        var token Token

        argument, text, err = parseExpression(text)
        if err != nil {
            return Expression{}, text, err
        }

        arguments = append(arguments, argument)
        token, text, err = nextToken(text)
        if token.kind == TokenCloseBrace {
            break
        }

        if token.kind != TokenComma {
            return Expression{}, text, errors.New("Missing comma")
        }
    }

    return Expression {
        kind: ExpressionFunction,
        function: function,
        arguments: arguments,
    }, text, err
}

func parseTerm(text string) (Expression, string, error) {
    token, text, err := nextToken(text)
    if err != nil {
        return Expression{}, text, err
    }

    switch token.kind {
    case TokenName:
        return parseFunction(token.name, text)
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
    case TokenRange:
        return Expression {
            kind: ExpressionRange,
            row: token.row,
            column: token.column,
            end_row: token.end_row,
            end_column: token.end_column,
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

    should_continue := true
    for should_continue {
        token, next_text, err := nextToken(text)
        if err != nil {
            return Expression{}, text, err
        }

        switch token.kind {
        case TokenAdd:
            result, text, err = parseOperation(result, ExpressionAdd, next_text)
            if err != nil {
                return Expression{}, text, err
            }
        default:
            should_continue = false
        }
    }

    return result, text, nil
}

