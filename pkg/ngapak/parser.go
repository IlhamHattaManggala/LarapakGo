package ngapak

import "fmt"

// Parser transforms a stream of tokens into an AST.
type Parser struct {
	tokens   []Token
	filename string
	pos      int
}

// NewParser creates a new Parser instance.
func NewParser(tokens []Token, filename string) *Parser {
	if filename == "" {
		filename = "<stdin>"
	}
	return &Parser{tokens: tokens, filename: filename, pos: 0}
}

func (p *Parser) current() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{Type: T_EOF}
}

func (p *Parser) check(t TokenType) bool {
	return p.current().Type == t
}

func (p *Parser) advance() Token {
	tok := p.current()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

func (p *Parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) consume(t TokenType, errMsg string) (Token, error) {
	if p.check(t) {
		return p.advance(), nil
	}
	cur := p.current()
	return Token{}, fmt.Errorf("%s:%d:%d: kleru sintaksis: %s (ditemokake: %s)", p.filename, cur.Line, cur.Column, errMsg, cur.Type)
}

// Parse parses the whole token stream into a ProgramNode.
func (p *Parser) Parse() (*ProgramNode, error) {
	var stmts []Statement

	// Skip leading newlines
	for p.match(T_NEWLINE) {
	}

	for !p.check(T_EOF) {
		stmt, err := p.statement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			stmts = append(stmts, stmt)
		}

		// Require newline or EOF between statements, or closing blocks
		if !p.check(T_EOF) && !p.check(T_NEWLINE) && p.pos > 0 && p.tokens[p.pos-1].Type != T_NEWLINE {
			if !p.check(T_RAMPUNG) && !p.check(T_LIYANE) {
				cur := p.current()
				return nil, fmt.Errorf("%s:%d:%d: kudu ana baris anyar sawise pranyatan", p.filename, cur.Line, cur.Column)
			}
		}

		for p.match(T_NEWLINE) {
		}
	}

	return &ProgramNode{Statements: stmts, Line: 1, Col: 1}, nil
}

func (p *Parser) statement() (Statement, error) {
	startTok := p.current()

	// 1. tulis <expr>
	if p.match(T_TULIS) {
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		return &TulisNode{Expr: expr, Line: startTok.Line, Col: startTok.Column}, nil
	}

	// 2. nek <cond> ya ... [liyane ...] rampung
	if p.match(T_NEK) {
		return p.ifStatement(startTok)
	}

	// 3. baleni <var> saka <start> nganti <end> ... rampung / baleni <var> neng <coll> ... rampung
	if p.match(T_BALENI) {
		return p.forStatement(startTok)
	}

	// 4. gawe <name>(...) ... rampung
	if p.match(T_GAWE) {
		return p.funcDefStatement(startTok)
	}

	// 5. balekna [<expr>]
	if p.match(T_BALEKNA) {
		var expr Expression
		var err error
		if !p.check(T_NEWLINE) && !p.check(T_EOF) && !p.check(T_RAMPUNG) && !p.check(T_LIYANE) {
			expr, err = p.expression()
			if err != nil {
				return nil, err
			}
		}
		return &ReturnNode{Expr: expr, Line: startTok.Line, Col: startTok.Column}, nil
	}

	// 6. mandeg (break)
	if p.match(T_MANDEG) {
		return &BreakNode{Line: startTok.Line, Col: startTok.Column}, nil
	}

	// 7. terusna (continue)
	if p.match(T_TERUSNA) {
		return &ContinueNode{Line: startTok.Line, Col: startTok.Column}, nil
	}

	// 8. Expression statement or Assignment
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	if p.match(T_ASSIGN) {
		valExpr, err := p.expression()
		if err != nil {
			return nil, err
		}

		switch target := expr.(type) {
		case *VarAccessNode:
			return &VarAssignNode{Name: target.Name, Expr: valExpr, Line: startTok.Line, Col: startTok.Column}, nil
		case *MemberAccessNode:
			return &MemberAssignNode{Obj: target.Obj, Member: target.Member, Expr: valExpr, Line: startTok.Line, Col: startTok.Column}, nil
		case *IndexAccessNode:
			return &IndexAssignNode{Obj: target.Obj, Index: target.Index, Expr: valExpr, Line: startTok.Line, Col: startTok.Column}, nil
		default:
			return nil, fmt.Errorf("%s:%d:%d: sisih kiwa penugasan ora sah", p.filename, startTok.Line, startTok.Column)
		}
	}

	return &ExprStatement{Expr: expr, Line: startTok.Line, Col: startTok.Column}, nil
}

func (p *Parser) ifStatement(startTok Token) (Statement, error) {
	cond, err := p.expression()
	if err != nil {
		return nil, err
	}

	if _, err := p.consume(T_YA, "dikarepake tembung kunci 'ya' sawise kondisi 'nek'"); err != nil {
		return nil, err
	}

	for p.match(T_NEWLINE) {
	}

	var trueBranch []Statement
	for !p.check(T_EOF) && !p.check(T_LIYANE) && !p.check(T_RAMPUNG) {
		stmt, err := p.statement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			trueBranch = append(trueBranch, stmt)
		}
		for p.match(T_NEWLINE) {
		}
	}

	var falseBranch []Statement
	if p.match(T_LIYANE) {
		for p.match(T_NEWLINE) {
		}
		for !p.check(T_EOF) && !p.check(T_RAMPUNG) {
			stmt, err := p.statement()
			if err != nil {
				return nil, err
			}
			if stmt != nil {
				falseBranch = append(falseBranch, stmt)
			}
			for p.match(T_NEWLINE) {
			}
		}
	}

	if _, err := p.consume(T_RAMPUNG, "dikarepake tembung kunci 'rampung' kanggo nutup blok 'nek'"); err != nil {
		return nil, err
	}

	return &IfNode{Condition: cond, TrueBranch: trueBranch, FalseBranch: falseBranch, Line: startTok.Line, Col: startTok.Column}, nil
}

func (p *Parser) forStatement(startTok Token) (Statement, error) {
	varTok, err := p.consume(T_IDENTIFIER, "dikarepake jeneng variabel sawise 'baleni'")
	if err != nil {
		return nil, err
	}
	varName := varTok.Value.(string)

	if p.match(T_NENG) {
		collExpr, err := p.expression()
		if err != nil {
			return nil, err
		}
		for p.match(T_NEWLINE) {
		}
		var body []Statement
		for !p.check(T_EOF) && !p.check(T_RAMPUNG) {
			stmt, err := p.statement()
			if err != nil {
				return nil, err
			}
			if stmt != nil {
				body = append(body, stmt)
			}
			for p.match(T_NEWLINE) {
			}
		}
		if _, err := p.consume(T_RAMPUNG, "dikarepake tembung kunci 'rampung' kanggo nutup blok 'baleni'"); err != nil {
			return nil, err
		}
		return &ForEachNode{VarName: varName, Collection: collExpr, Body: body, Line: startTok.Line, Col: startTok.Column}, nil
	}

	if _, err := p.consume(T_SAKA, "dikarepake tembung kunci 'saka' utawa 'neng' sawise variabel loop"); err != nil {
		return nil, err
	}
	startExpr, err := p.expression()
	if err != nil {
		return nil, err
	}

	if _, err := p.consume(T_NGANTI, "dikarepake tembung kunci 'nganti' sawise nilai wiwitan loop"); err != nil {
		return nil, err
	}
	endExpr, err := p.expression()
	if err != nil {
		return nil, err
	}

	for p.match(T_NEWLINE) {
	}

	var body []Statement
	for !p.check(T_EOF) && !p.check(T_RAMPUNG) {
		stmt, err := p.statement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			body = append(body, stmt)
		}
		for p.match(T_NEWLINE) {
		}
	}

	if _, err := p.consume(T_RAMPUNG, "dikarepake tembung kunci 'rampung' kanggo nutup blok 'baleni'"); err != nil {
		return nil, err
	}

	return &ForNode{VarName: varName, StartExpr: startExpr, EndExpr: endExpr, Body: body, Line: startTok.Line, Col: startTok.Column}, nil
}

func (p *Parser) funcDefStatement(startTok Token) (Statement, error) {
	name := "anonymous"
	if p.check(T_IDENTIFIER) {
		nameTok := p.advance()
		name = nameTok.Value.(string)
	}

	if _, err := p.consume(T_LPAREN, "dikarepake '(' sawise jeneng fungsi"); err != nil {
		return nil, err
	}

	var params []string
	if !p.check(T_RPAREN) {
		paramTok, err := p.consume(T_IDENTIFIER, "dikarepake jeneng parameter")
		if err != nil {
			return nil, err
		}
		params = append(params, paramTok.Value.(string))
		for p.match(T_COMMA) {
			paramTok, err := p.consume(T_IDENTIFIER, "dikarepake jeneng parameter sawise koma")
			if err != nil {
				return nil, err
			}
			params = append(params, paramTok.Value.(string))
		}
	}

	if _, err := p.consume(T_RPAREN, "dikarepake ')' sawise dhaptar parameter"); err != nil {
		return nil, err
	}

	for p.match(T_NEWLINE) {
	}

	var body []Statement
	for !p.check(T_EOF) && !p.check(T_RAMPUNG) {
		stmt, err := p.statement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			body = append(body, stmt)
		}
		for p.match(T_NEWLINE) {
		}
	}

	if _, err := p.consume(T_RAMPUNG, "dikarepake tembung kunci 'rampung' kanggo nutup fungsi"); err != nil {
		return nil, err
	}

	return &FuncDefNode{Name: name, Params: params, Body: body, Line: startTok.Line, Col: startTok.Column}, nil
}

func (p *Parser) expression() (Expression, error) {
	return p.logicalOr()
}

func (p *Parser) logicalOr() (Expression, error) {
	expr, err := p.logicalAnd()
	if err != nil {
		return nil, err
	}

	for p.match(T_UTAWA) {
		opTok := p.tokens[p.pos-1]
		right, err := p.logicalAnd()
		if err != nil {
			return nil, err
		}
		expr = &BinOpNode{Left: expr, Op: opTok.Type, Right: right, Line: opTok.Line, Col: opTok.Column}
	}
	return expr, nil
}

func (p *Parser) logicalAnd() (Expression, error) {
	expr, err := p.comparison()
	if err != nil {
		return nil, err
	}

	for p.match(T_LAN) {
		opTok := p.tokens[p.pos-1]
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}
		expr = &BinOpNode{Left: expr, Op: opTok.Type, Right: right, Line: opTok.Line, Col: opTok.Column}
	}
	return expr, nil
}

func (p *Parser) comparison() (Expression, error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.match(T_EQ, T_NEQ, T_LT, T_GT, T_LTE, T_GTE) {
		opTok := p.tokens[p.pos-1]
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		expr = &BinOpNode{Left: expr, Op: opTok.Type, Right: right, Line: opTok.Line, Col: opTok.Column}
	}
	return expr, nil
}

func (p *Parser) term() (Expression, error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.match(T_PLUS, T_MINUS) {
		opTok := p.tokens[p.pos-1]
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		expr = &BinOpNode{Left: expr, Op: opTok.Type, Right: right, Line: opTok.Line, Col: opTok.Column}
	}
	return expr, nil
}

func (p *Parser) factor() (Expression, error) {
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}

	for p.match(T_MUL, T_DIV) {
		opTok := p.tokens[p.pos-1]
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		expr = &BinOpNode{Left: expr, Op: opTok.Type, Right: right, Line: opTok.Line, Col: opTok.Column}
	}
	return expr, nil
}

func (p *Parser) unary() (Expression, error) {
	if p.match(T_ORA, T_MINUS) {
		opTok := p.tokens[p.pos-1]
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		return &UnaryOpNode{Op: opTok.Type, Expr: right, Line: opTok.Line, Col: opTok.Column}, nil
	}
	return p.primary()
}

func (p *Parser) primary() (Expression, error) {
	expr, err := p.basePrimary()
	if err != nil {
		return nil, err
	}

	for {
		if p.match(T_DOT) {
			dotTok := p.tokens[p.pos-1]
			// Member or method identifier
			if !p.check(T_IDENTIFIER) && !p.isKeyword(p.current().Type) {
				return nil, fmt.Errorf("%s:%d:%d: dikarepake jeneng properti utawa metode sawise '.'", p.filename, dotTok.Line, dotTok.Column)
			}
			nameTok := p.advance()
			name := fmt.Sprintf("%v", nameTok.Value)

			if p.match(T_LPAREN) {
				var args []Expression
				for p.match(T_NEWLINE) {
				}
				if !p.check(T_RPAREN) {
					arg, err := p.expression()
					if err != nil {
						return nil, err
					}
					args = append(args, arg)
					for {
						for p.match(T_NEWLINE) {
						}
						if !p.match(T_COMMA) {
							break
						}
						for p.match(T_NEWLINE) {
						}
						if p.check(T_RPAREN) {
							break
						}
						arg, err := p.expression()
						if err != nil {
							return nil, err
						}
						args = append(args, arg)
					}
				}
				for p.match(T_NEWLINE) {
				}
				if _, err := p.consume(T_RPAREN, "dikarepake ')' sawise argumen metode"); err != nil {
					return nil, err
				}
				expr = &MethodCallNode{Obj: expr, Method: name, Args: args, Line: dotTok.Line, Col: dotTok.Column}
			} else {
				expr = &MemberAccessNode{Obj: expr, Member: name, Line: dotTok.Line, Col: dotTok.Column}
			}
		} else if p.match(T_LBRACKET) {
			bracketTok := p.tokens[p.pos-1]
			idxExpr, err := p.expression()
			if err != nil {
				return nil, err
			}
			if _, err := p.consume(T_RBRACKET, "dikarepake ']' sawise indeks"); err != nil {
				return nil, err
			}
			expr = &IndexAccessNode{Obj: expr, Index: idxExpr, Line: bracketTok.Line, Col: bracketTok.Column}
		} else {
			break
		}
	}

	return expr, nil
}

func (p *Parser) isKeyword(t TokenType) bool {
	return t == T_TULIS || t == T_NEK || t == T_YA || t == T_LIYANE || t == T_RAMPUNG ||
		t == T_BALENI || t == T_SAKA || t == T_NGANTI || t == T_GAWE || t == T_BALEKNA ||
		t == T_BENER || t == T_SALAH || t == T_KOSONG || t == T_KELAS || t == T_EXTENDS ||
		t == T_LAN || t == T_UTAWA || t == T_ORA || t == T_NENG || t == T_MANDEG || t == T_TERUSNA
}

func (p *Parser) basePrimary() (Expression, error) {
	startTok := p.current()

	if p.match(T_BENER) {
		return &LiteralNode{Value: true, Line: startTok.Line, Col: startTok.Column}, nil
	}
	if p.match(T_SALAH) {
		return &LiteralNode{Value: false, Line: startTok.Line, Col: startTok.Column}, nil
	}
	if p.match(T_KOSONG) {
		return &LiteralNode{Value: nil, Line: startTok.Line, Col: startTok.Column}, nil
	}
	if p.match(T_NUMBER, T_STRING) {
		return &LiteralNode{Value: p.tokens[p.pos-1].Value, Line: startTok.Line, Col: startTok.Column}, nil
	}

	// { key: val, ... } Dictionary literal
	if p.match(T_LBRACE) {
		for p.match(T_NEWLINE) {
		}
		if p.match(T_RBRACE) {
			return &DictNode{Keys: nil, Values: nil, Line: startTok.Line, Col: startTok.Column}, nil
		}

		firstKey, err := p.expression()
		if err != nil {
			return nil, err
		}
		for p.match(T_NEWLINE) {
		}
		if _, err := p.consume(T_COLON, "dikarepake ':' sawise kunci kamus"); err != nil {
			return nil, err
		}
		for p.match(T_NEWLINE) {
		}
		firstVal, err := p.expression()
		if err != nil {
			return nil, err
		}
		keys := []Expression{firstKey}
		values := []Expression{firstVal}

		for {
			for p.match(T_NEWLINE) {
			}
			if !p.match(T_COMMA) {
				break
			}
			for p.match(T_NEWLINE) {
			}
			if p.check(T_RBRACE) {
				break
			}
			k, err := p.expression()
			if err != nil {
				return nil, err
			}
			for p.match(T_NEWLINE) {
			}
			if _, err := p.consume(T_COLON, "dikarepake ':' sawise kunci kamus"); err != nil {
				return nil, err
			}
			for p.match(T_NEWLINE) {
			}
			v, err := p.expression()
			if err != nil {
				return nil, err
			}
			keys = append(keys, k)
			values = append(values, v)
		}
		for p.match(T_NEWLINE) {
		}
		if _, err := p.consume(T_RBRACE, "dikarepake '}' kanggo nutup kamus"); err != nil {
			return nil, err
		}
		return &DictNode{Keys: keys, Values: values, Line: startTok.Line, Col: startTok.Column}, nil
	}

	// [key: val, ...] or [elem1, elem2, ...]
	if p.match(T_LBRACKET) {
		for p.match(T_NEWLINE) {
		}
		if p.match(T_RBRACKET) {
			return &ListNode{Elements: nil, Line: startTok.Line, Col: startTok.Column}, nil
		}

		firstExpr, err := p.expression()
		if err != nil {
			return nil, err
		}
		for p.match(T_NEWLINE) {
		}

		if p.match(T_COLON) {
			for p.match(T_NEWLINE) {
			}
			// Dictionary literal
			valExpr, err := p.expression()
			if err != nil {
				return nil, err
			}
			keys := []Expression{firstExpr}
			values := []Expression{valExpr}

			for {
				for p.match(T_NEWLINE) {
				}
				if !p.match(T_COMMA) {
					break
				}
				for p.match(T_NEWLINE) {
				}
				if p.check(T_RBRACKET) {
					break
				}
				k, err := p.expression()
				if err != nil {
					return nil, err
				}
				for p.match(T_NEWLINE) {
				}
				if _, err := p.consume(T_COLON, "dikarepake ':' sawise kunci kamus"); err != nil {
					return nil, err
				}
				for p.match(T_NEWLINE) {
				}
				v, err := p.expression()
				if err != nil {
					return nil, err
				}
				keys = append(keys, k)
				values = append(values, v)
			}
			for p.match(T_NEWLINE) {
			}
			if _, err := p.consume(T_RBRACKET, "dikarepake ']' kanggo nutup kamus"); err != nil {
				return nil, err
			}
			return &DictNode{Keys: keys, Values: values, Line: startTok.Line, Col: startTok.Column}, nil
		}

		// List literal
		elements := []Expression{firstExpr}
		for {
			for p.match(T_NEWLINE) {
			}
			if !p.match(T_COMMA) {
				break
			}
			for p.match(T_NEWLINE) {
			}
			if p.check(T_RBRACKET) {
				break
			}
			elem, err := p.expression()
			if err != nil {
				return nil, err
			}
			elements = append(elements, elem)
		}
		for p.match(T_NEWLINE) {
		}
		if _, err := p.consume(T_RBRACKET, "dikarepake ']' kanggo nutup dhaptar"); err != nil {
			return nil, err
		}
		return &ListNode{Elements: elements, Line: startTok.Line, Col: startTok.Column}, nil
	}

	// Function definition as expression: gawe(...) ... rampung
	if p.match(T_GAWE) {
		fStmt, err := p.funcDefStatement(startTok)
		if err != nil {
			return nil, err
		}
		return fStmt.(*FuncDefNode), nil
	}

	// Identifier or function call: name(...)
	if p.check(T_IDENTIFIER) {
		idTok := p.advance()
		name := idTok.Value.(string)

		if p.match(T_LPAREN) {
			var args []Expression
			for p.match(T_NEWLINE) {
			}
			if !p.check(T_RPAREN) {
				arg, err := p.expression()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				for {
					for p.match(T_NEWLINE) {
					}
					if !p.match(T_COMMA) {
						break
					}
					for p.match(T_NEWLINE) {
					}
					if p.check(T_RPAREN) {
						break
					}
					arg, err := p.expression()
					if err != nil {
						return nil, err
					}
					args = append(args, arg)
				}
			}
			for p.match(T_NEWLINE) {
			}
			if _, err := p.consume(T_RPAREN, "dikarepake ')' sawise dhaptar argumen"); err != nil {
				return nil, err
			}
			return &FuncCallNode{Callee: name, Args: args, Line: startTok.Line, Col: startTok.Column}, nil
		}
		return &VarAccessNode{Name: name, Line: startTok.Line, Col: startTok.Column}, nil
	}

	// Grouping (expr)
	if p.match(T_LPAREN) {
		for p.match(T_NEWLINE) {
		}
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		for p.match(T_NEWLINE) {
		}
		if _, err := p.consume(T_RPAREN, "dikarepake ')' sawise ekspresi"); err != nil {
			return nil, err
		}
		return expr, nil
	}

	cur := p.current()
	return nil, fmt.Errorf("%s:%d:%d: sintaksis ora sah ing token %q", p.filename, cur.Line, cur.Column, cur.Value)
}

