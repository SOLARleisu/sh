// Copyright (c) 2017, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package expand

import (
	"context"
	"fmt"
	"strconv"

	"mvdan.cc/sh/syntax"
)

func (c *Context) ExpandArithm(ctx context.Context, expr syntax.ArithmExpr) int {
	switch x := expr.(type) {
	case *syntax.Word:
		str := c.ExpandLiteral(ctx, x)
		// recursively fetch vars
		for str != "" {
			val := c.envGet(str)
			if val == "" {
				break
			}
			str = val
		}
		// default to 0
		return atoi(str)
	case *syntax.ParenArithm:
		return c.ExpandArithm(ctx, x.X)
	case *syntax.UnaryArithm:
		switch x.Op {
		case syntax.Inc, syntax.Dec:
			name := x.X.(*syntax.Word).Parts[0].(*syntax.Lit).Value
			old := atoi(c.envGet(name))
			val := old
			if x.Op == syntax.Inc {
				val++
			} else {
				val--
			}
			c.envSet(name, strconv.Itoa(val))
			if x.Post {
				return old
			}
			return val
		}
		val := c.ExpandArithm(ctx, x.X)
		switch x.Op {
		case syntax.Not:
			return oneIf(val == 0)
		case syntax.Plus:
			return val
		default: // syntax.Minus
			return -val
		}
	case *syntax.BinaryArithm:
		switch x.Op {
		case syntax.Assgn, syntax.AddAssgn, syntax.SubAssgn,
			syntax.MulAssgn, syntax.QuoAssgn, syntax.RemAssgn,
			syntax.AndAssgn, syntax.OrAssgn, syntax.XorAssgn,
			syntax.ShlAssgn, syntax.ShrAssgn:
			return c.assgnArit(ctx, x)
		case syntax.Quest: // Colon can't happen here
			cond := c.ExpandArithm(ctx, x.X)
			b2 := x.Y.(*syntax.BinaryArithm) // must have Op==Colon
			if cond == 1 {
				return c.ExpandArithm(ctx, b2.X)
			}
			return c.ExpandArithm(ctx, b2.Y)
		}
		return binArit(x.Op, c.ExpandArithm(ctx, x.X), c.ExpandArithm(ctx, x.Y))
	default:
		panic(fmt.Sprintf("unexpected arithm expr: %T", x))
	}
}

func oneIf(b bool) int {
	if b {
		return 1
	}
	return 0
}

// atoi is just a shorthand for strconv.Atoi that ignores the error,
// just like shells do.
func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func (c *Context) assgnArit(ctx context.Context, b *syntax.BinaryArithm) int {
	name := b.X.(*syntax.Word).Parts[0].(*syntax.Lit).Value
	val := atoi(c.envGet(name))
	arg := c.ExpandArithm(ctx, b.Y)
	switch b.Op {
	case syntax.Assgn:
		val = arg
	case syntax.AddAssgn:
		val += arg
	case syntax.SubAssgn:
		val -= arg
	case syntax.MulAssgn:
		val *= arg
	case syntax.QuoAssgn:
		val /= arg
	case syntax.RemAssgn:
		val %= arg
	case syntax.AndAssgn:
		val &= arg
	case syntax.OrAssgn:
		val |= arg
	case syntax.XorAssgn:
		val ^= arg
	case syntax.ShlAssgn:
		val <<= uint(arg)
	case syntax.ShrAssgn:
		val >>= uint(arg)
	}
	c.envSet(name, strconv.Itoa(val))
	return val
}

func intPow(a, b int) int {
	p := 1
	for b > 0 {
		if b&1 != 0 {
			p *= a
		}
		b >>= 1
		a *= a
	}
	return p
}

func binArit(op syntax.BinAritOperator, x, y int) int {
	switch op {
	case syntax.Add:
		return x + y
	case syntax.Sub:
		return x - y
	case syntax.Mul:
		return x * y
	case syntax.Quo:
		return x / y
	case syntax.Rem:
		return x % y
	case syntax.Pow:
		return intPow(x, y)
	case syntax.Eql:
		return oneIf(x == y)
	case syntax.Gtr:
		return oneIf(x > y)
	case syntax.Lss:
		return oneIf(x < y)
	case syntax.Neq:
		return oneIf(x != y)
	case syntax.Leq:
		return oneIf(x <= y)
	case syntax.Geq:
		return oneIf(x >= y)
	case syntax.And:
		return x & y
	case syntax.Or:
		return x | y
	case syntax.Xor:
		return x ^ y
	case syntax.Shr:
		return x >> uint(y)
	case syntax.Shl:
		return x << uint(y)
	case syntax.AndArit:
		return oneIf(x != 0 && y != 0)
	case syntax.OrArit:
		return oneIf(x != 0 || y != 0)
	default: // syntax.Comma
		// x is executed but its result discarded
		return y
	}
}
