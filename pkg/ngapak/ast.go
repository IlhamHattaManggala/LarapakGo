package ngapak

// Node represents an AST node.
type Node interface {
	Pos() (int, int)
}

// Statement represents a statement node.
type Statement interface {
	Node
	statementNode()
}

// Expression represents an expression node.
type Expression interface {
	Node
	expressionNode()
}

// ProgramNode holds all top-level statements.
type ProgramNode struct {
	Statements []Statement
	Line       int
	Col        int
}

func (p *ProgramNode) Pos() (int, int) { return p.Line, p.Col }

// TulisNode represents a 'tulis <expr>' statement.
type TulisNode struct {
	Expr Expression
	Line int
	Col  int
}

func (n *TulisNode) Pos() (int, int) { return n.Line, n.Col }
func (n *TulisNode) statementNode()  {}

// VarAssignNode represents 'varName = <expr>'.
type VarAssignNode struct {
	Name string
	Expr Expression
	Line int
	Col  int
}

func (n *VarAssignNode) Pos() (int, int) { return n.Line, n.Col }
func (n *VarAssignNode) statementNode()  {}

// MemberAssignNode represents 'obj.field = <expr>'.
type MemberAssignNode struct {
	Obj    Expression
	Member string
	Expr   Expression
	Line   int
	Col    int
}

func (n *MemberAssignNode) Pos() (int, int) { return n.Line, n.Col }
func (n *MemberAssignNode) statementNode()  {}

// IndexAssignNode represents 'obj[idx] = <expr>'.
type IndexAssignNode struct {
	Obj   Expression
	Index Expression
	Expr  Expression
	Line  int
	Col   int
}

func (n *IndexAssignNode) Pos() (int, int) { return n.Line, n.Col }
func (n *IndexAssignNode) statementNode()  {}

// IfNode represents 'nek <cond> ya ... [liyane ...] rampung'.
type IfNode struct {
	Condition   Expression
	TrueBranch  []Statement
	FalseBranch []Statement
	Line        int
	Col         int
}

func (n *IfNode) Pos() (int, int) { return n.Line, n.Col }
func (n *IfNode) statementNode()  {}

// ForNode represents 'baleni <var> saka <start> nganti <end> ... rampung'.
type ForNode struct {
	VarName   string
	StartExpr Expression
	EndExpr   Expression
	Body      []Statement
	Line      int
	Col       int
}

func (n *ForNode) Pos() (int, int) { return n.Line, n.Col }
func (n *ForNode) statementNode()  {}

// ForEachNode represents 'baleni <var> neng <coll> ... rampung'.
type ForEachNode struct {
	VarName    string
	Collection Expression
	Body       []Statement
	Line       int
	Col        int
}

func (n *ForEachNode) Pos() (int, int) { return n.Line, n.Col }
func (n *ForEachNode) statementNode()  {}

// FuncDefNode represents 'gawe <name>(param1, ...) ... rampung'.
type FuncDefNode struct {
	Name   string
	Params []string
	Body   []Statement
	Line   int
	Col    int
}

func (n *FuncDefNode) Pos() (int, int) { return n.Line, n.Col }
func (n *FuncDefNode) statementNode()  {}
func (n *FuncDefNode) expressionNode() {}

// ReturnNode represents 'balekna [<expr>]'.
type ReturnNode struct {
	Expr Expression
	Line int
	Col  int
}

func (n *ReturnNode) Pos() (int, int) { return n.Line, n.Col }
func (n *ReturnNode) statementNode()  {}

// BreakNode represents 'mandeg'.
type BreakNode struct {
	Line int
	Col  int
}

func (n *BreakNode) Pos() (int, int) { return n.Line, n.Col }
func (n *BreakNode) statementNode()  {}

// ContinueNode represents 'terusna'.
type ContinueNode struct {
	Line int
	Col  int
}

func (n *ContinueNode) Pos() (int, int) { return n.Line, n.Col }
func (n *ContinueNode) statementNode()  {}

// ExprStatement wraps an Expression used as a Statement.
type ExprStatement struct {
	Expr Expression
	Line int
	Col  int
}

func (n *ExprStatement) Pos() (int, int) { return n.Line, n.Col }
func (n *ExprStatement) statementNode()  {}

// LiteralNode represents a constant value (int64, float64, string, bool, nil).
type LiteralNode struct {
	Value any
	Line  int
	Col   int
}

func (n *LiteralNode) Pos() (int, int) { return n.Line, n.Col }
func (n *LiteralNode) expressionNode() {}

// VarAccessNode represents variable lookup by name.
type VarAccessNode struct {
	Name string
	Line int
	Col  int
}

func (n *VarAccessNode) Pos() (int, int) { return n.Line, n.Col }
func (n *VarAccessNode) expressionNode() {}

// BinOpNode represents binary operations (e.g. a + b, x == y, a lan b).
type BinOpNode struct {
	Left  Expression
	Op    TokenType
	Right Expression
	Line  int
	Col   int
}

func (n *BinOpNode) Pos() (int, int) { return n.Line, n.Col }
func (n *BinOpNode) expressionNode() {}

// UnaryOpNode represents unary operations (e.g. -x, ora y).
type UnaryOpNode struct {
	Op   TokenType
	Expr Expression
	Line int
	Col  int
}

func (n *UnaryOpNode) Pos() (int, int) { return n.Line, n.Col }
func (n *UnaryOpNode) expressionNode() {}

// FuncCallNode represents 'callee(args...)'.
type FuncCallNode struct {
	Callee string
	Args   []Expression
	Line   int
	Col    int
}

func (n *FuncCallNode) Pos() (int, int) { return n.Line, n.Col }
func (n *FuncCallNode) expressionNode() {}

// MemberAccessNode represents 'obj.field'.
type MemberAccessNode struct {
	Obj    Expression
	Member string
	Line   int
	Col    int
}

func (n *MemberAccessNode) Pos() (int, int) { return n.Line, n.Col }
func (n *MemberAccessNode) expressionNode() {}

// MethodCallNode represents 'obj.method(args...)'.
type MethodCallNode struct {
	Obj    Expression
	Method string
	Args   []Expression
	Line   int
	Col    int
}

func (n *MethodCallNode) Pos() (int, int) { return n.Line, n.Col }
func (n *MethodCallNode) expressionNode() {}

// IndexAccessNode represents 'obj[index]'.
type IndexAccessNode struct {
	Obj   Expression
	Index Expression
	Line  int
	Col   int
}

func (n *IndexAccessNode) Pos() (int, int) { return n.Line, n.Col }
func (n *IndexAccessNode) expressionNode() {}

// DictNode represents dictionary literal '[key: val, key2: val2]'.
type DictNode struct {
	Keys   []Expression
	Values []Expression
	Line   int
	Col    int
}

func (n *DictNode) Pos() (int, int) { return n.Line, n.Col }
func (n *DictNode) expressionNode() {}

// ListNode represents list literal '[elem1, elem2]'.
type ListNode struct {
	Elements []Expression
	Line     int
	Col      int
}

func (n *ListNode) Pos() (int, int) { return n.Line, n.Col }
func (n *ListNode) expressionNode() {}

