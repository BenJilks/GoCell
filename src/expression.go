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
    TokenConstant
    TokenRange
    TokenEmpty
)

type Range struct {
    start CellPosition
    end CellPosition
}

func (r Range) Shift(offset CellPosition) Range {
    return Range {
        CellPosition {
            r.start.row + offset.row,
            r.start.column + offset.column,
        },
        CellPosition {
            r.end.row + offset.row,
            r.end.column + offset.column,
        },
    }
}

type Token struct {
    kind TokenKind
    name string
    number float64
    position CellPosition
    cellRange Range
}

type ExpressionKind int
const (
    ExpressionAdd ExpressionKind = iota
    ExpressionNumber
    ExpressionCell
    ExpressionConstant
    ExpressionRange
    ExpressionFunction
)

type Expression struct {
    kind ExpressionKind
    number float64

    lhs *Expression
    rhs *Expression

    position CellPosition
    cellRange Range

    function string
    arguments []*Expression
}

const BlockSize = 1024;

type ExpressionAllocator struct {
    blocks [][]Expression
    used_in_current_block int
}

func newExpressionAllocator() ExpressionAllocator {
    blocks := make([][]Expression, 1)
    blocks = append(blocks, make([]Expression, BlockSize))
    return ExpressionAllocator {
        blocks: blocks,
        used_in_current_block: 0,
    }
}

func (allocator *ExpressionAllocator) New() *Expression {
    if allocator.used_in_current_block >= BlockSize {
        allocator.used_in_current_block = 0
        allocator.blocks = append(allocator.blocks, make([]Expression, BlockSize))
    }
    
    index := allocator.used_in_current_block
    block := allocator.blocks[len(allocator.blocks)-1]

    allocator.used_in_current_block += 1
    return &block[index]
}

func isLetter(c byte) bool {
    return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func isDigit(c byte) bool {
    return c >= '0' && c <= '9'
}

func isDirection(c byte) bool {
    return c == '^' || c == '>' || c == 'v' || c == '<'
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

func parseRange(start CellPosition, text string, position CellPosition) (Token, string, error) {
    var end Token
    var err error

    if isDirection(text[0]) {
        end, text, err = parseRelativeCellReferance(text, position)
    } else {
        end, text, err = parseCellReference(text, position)
    }

    if err != nil {
        return Token{}, text, err
    }

    return Token {
        kind: TokenRange,
        cellRange: Range {
            start: start,
            end: end.position,
        },
    }, text, nil
}

func parseInt(text string) (int, string, error) {
    i := 0
    for i < len(text) && isDigit(text[i]) {
        i += 1
    }

    result, err := strconv.ParseInt(text[0:i], 10, 32)
    if err != nil {
        return -1, text, err
    }

    return int(result), text[i:], nil
}

func parseRelativeCellReferance(text string, position CellPosition) (Token, string, error) {
    var err error

    direction := DirectionNone
    switch text[0] {
    case '^': direction = DirectionUp
    case '>': direction = DirectionRight
    case 'v': direction = DirectionDown
    case '<': direction = DirectionLeft
    default:
        panic(0)
    }

    text = text[1:]
    offset := 1
    if len(text) > 0 && isDigit(text[0]) {
        offset, text, err = parseInt(text)
        if err != nil {
            return Token{}, text, err
        }
    }

    refPosition := position.Offset(direction, offset)
    return Token { kind: TokenCell, position: refPosition }, text, nil
}

func parseConstantReferance(text string, position CellPosition) (Token, string, error) {
    ref, text, err := parseCellReference(text[1:], position)
    if err != nil {
        return Token{}, text, err
    }

    ref.kind = TokenConstant
    return ref, text, nil
}

func parseCellReference(text string, position CellPosition) (Token, string, error) {
    column_name := strings.ToUpper(text)[0]
    column := column_name - 'A'
    row, text, err := parseInt(text[1:])
    if err != nil {
        return Token{}, text, err
    }

    refPosition := CellPosition { int(row) - 1, int(column) }
    if len(text) > 0 && text[0] == ':' {
        return parseRange(refPosition, text[1:], position)
    }

    return Token { kind: TokenCell, position: refPosition }, text, nil
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

func nextToken(text string, position CellPosition) (Token, string, error) {
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
    case c == '$':
        return parseConstantReferance(text, position)
    case isDirection(c):
        return parseRelativeCellReferance(text, position)
    case isLetter(c):
        if name, text, err := parseName(text); err == nil {
            return name, text, err
        } else {
            return parseCellReference(text, position)
        }
    case isDigit(c):
        return parseNumber(text)
    default:
        return Token {}, text[1:], fmt.Errorf(
            "Unexpected char '%c'", c)
    }
}

func expect(kind TokenKind, text string, position CellPosition) (string, error) {
    token, text, err := nextToken(text, position)
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

func parseFunction(allocator *ExpressionAllocator,
                   function string,
                   text string,
                   position CellPosition) (*Expression, string, error) {
    text, err := expect(TokenOpenBrace, text, position)
    if err != nil {
        return nil, text, err
    }
    
    arguments := make([]*Expression, 0)
    for {
        var argument *Expression
        var token Token

        argument, text, err = parseExpression(allocator, text, position)
        if err != nil {
            return nil, text, err
        }

        arguments = append(arguments, argument)
        token, text, err = nextToken(text, position)
        if token.kind == TokenCloseBrace {
            break
        }

        if token.kind != TokenComma {
            return nil, text, errors.New("Missing comma")
        }
    }

    expression := allocator.New()
    expression.kind = ExpressionFunction
    expression.function = function
    expression.arguments = arguments
    return expression, text, nil
}

func parseTerm(allocator *ExpressionAllocator,
               text string,
               position CellPosition) (*Expression, string, error) {
    token, text, err := nextToken(text, position)
    if err != nil {
        return nil, text, err
    }

    switch token.kind {
    case TokenName:
        return parseFunction(allocator, token.name, text, position)
    case TokenNumber:
        expression := allocator.New()
        expression.kind = ExpressionNumber
        expression.number = token.number
        return expression, text, nil
    case TokenCell:
        expression := allocator.New()
        expression.kind = ExpressionCell
        expression.position = token.position
        return expression, text, nil
    case TokenConstant:
        expression := allocator.New()
        expression.kind = ExpressionConstant
        expression.position = token.position
        return expression, text, nil
    case TokenRange:
        expression := allocator.New()
        expression.kind = ExpressionRange
        expression.cellRange = token.cellRange
        return expression, text, nil
    case TokenAdd:
        return nil, text, errors.New(
            "Unexpected '+', expected value")
    case TokenEmpty:
        return nil, text, errors.New(
            "Expected value, got nothing instead")
    default:
        panic(0)
    }
}

func parseOperation(allocator *ExpressionAllocator,
                    lhs *Expression,
                    kind ExpressionKind,
                    text string,
                    position CellPosition) (*Expression, string, error) {
    rhs, text, err := parseExpression(allocator, text, position)
    if err != nil {
        return nil, text, err
    }

    expression := allocator.New()
    expression.kind = kind
    expression.lhs = lhs
    expression.rhs = rhs
    return expression, text, nil
}

func parseExpression(allocator *ExpressionAllocator,
                     text string,
                     position CellPosition) (*Expression, string, error) {
    result, text, err := parseTerm(allocator, text, position)
    if err != nil {
        return nil, text, err
    }

    should_continue := true
    for should_continue {
        token, next_text, err := nextToken(text, position)
        if err != nil {
            return nil, text, err
        }

        switch token.kind {
        case TokenAdd:
            result, text, err = parseOperation(allocator, result, ExpressionAdd, next_text, position)
            if err != nil {
                return nil, text, err
            }
        default:
            should_continue = false
        }
    }

    return result, text, nil
}

