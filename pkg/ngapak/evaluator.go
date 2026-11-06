package ngapak

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Environment holds variable scopes in NgapakIn.
type Environment struct {
	values map[string]any
	parent *Environment
}

// NewEnvironment creates a new environment, optionally with a parent.
func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		values: make(map[string]any),
		parent: parent,
	}
}

// Get retrieves a variable value from the environment chain.
func (e *Environment) Get(name string) (any, bool) {
	if val, ok := e.values[name]; ok {
		return val, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

// Set defines or overrides a variable in the local scope.
func (e *Environment) Set(name string, val any) {
	e.values[name] = val
}

// Assign updates an existing variable in the chain or sets it in the local scope.
func (e *Environment) Assign(name string, val any) {
	if _, ok := e.values[name]; ok {
		e.values[name] = val
		return
	}
	if e.parent != nil {
		if e.parent.assignExisting(name, val) {
			return
		}
	}
	e.values[name] = val
}

func (e *Environment) assignExisting(name string, val any) bool {
	if _, ok := e.values[name]; ok {
		e.values[name] = val
		return true
	}
	if e.parent != nil {
		return e.parent.assignExisting(name, val)
	}
	return false
}

// Function represents a user-defined Ngapak function.
type Function struct {
	Name    string
	Params  []string
	Body    []Statement
	Closure *Environment
}

func (f *Function) String() string {
	return fmt.Sprintf("<gawe %s>", f.Name)
}

// BuiltinFunc represents a Go-backed function exposed to Ngapak.
type BuiltinFunc struct {
	Name string
	Fn   func(args []any) (any, error)
}

func (b *BuiltinFunc) String() string {
	return fmt.Sprintf("<fungsi bawaan %s>", b.Name)
}

// Signals for control flow.
type returnSignal struct {
	value any
}

type breakSignal struct{}
type continueSignal struct{}

// Evaluator executes the AST.
type Evaluator struct {
	Globals *Environment
	Output  io.Writer
}

// NewEvaluator creates a new Evaluator.
func NewEvaluator() *Evaluator {
	ev := &Evaluator{
		Globals: NewEnvironment(nil),
		Output:  os.Stdout,
	}
	ev.registerBuiltins()
	return ev
}

func (ev *Evaluator) registerBuiltins() {
	ev.Globals.Set("dowo", &BuiltinFunc{
		Name: "dowo",
		Fn: func(args []any) (any, error) {
			if len(args) == 0 || args[0] == nil {
				return int64(0), nil
			}
			switch v := args[0].(type) {
			case string:
				return int64(len(v)), nil
			case []any:
				return int64(len(v)), nil
			case map[string]any:
				return int64(len(v)), nil
			default:
				return int64(1), nil
			}
		},
	})

	ev.Globals.Set("angka", &BuiltinFunc{
		Name: "angka",
		Fn: func(args []any) (any, error) {
			if len(args) == 0 || args[0] == nil {
				return int64(0), nil
			}
			switch v := args[0].(type) {
			case int64:
				return v, nil
			case int:
				return int64(v), nil
			case float64:
				return v, nil
			case string:
				if strings.Contains(v, ".") {
					if f, err := strconv.ParseFloat(v, 64); err == nil {
						return f, nil
					}
				}
				if i, err := strconv.ParseInt(v, 10, 64); err == nil {
					return i, nil
				}
				return int64(0), nil
			default:
				return int64(0), nil
			}
		},
	})

	ev.Globals.Set("teks", &BuiltinFunc{
		Name: "teks",
		Fn: func(args []any) (any, error) {
			if len(args) == 0 || args[0] == nil {
				return "kosong", nil
			}
			return ev.formatValue(args[0]), nil
		},
	})
}

// ExecuteProgram executes a full parsed program.
func (ev *Evaluator) ExecuteProgram(prog *ProgramNode) (any, error) {
	return ev.executeBlock(prog.Statements, ev.Globals)
}

func (ev *Evaluator) executeBlock(stmts []Statement, env *Environment) (any, error) {
	for _, stmt := range stmts {
		val, err := ev.executeStatement(stmt, env)
		if err != nil {
			return nil, err
		}
		if ret, ok := val.(returnSignal); ok {
			return ret, nil
		}
		if _, ok := val.(breakSignal); ok {
			return val, nil
		}
		if _, ok := val.(continueSignal); ok {
			return val, nil
		}
	}
	return nil, nil
}

func (ev *Evaluator) executeStatement(stmt Statement, env *Environment) (any, error) {
	switch s := stmt.(type) {
	case *TulisNode:
		val, err := ev.evaluateExpr(s.Expr, env)
		if err != nil {
			return nil, err
		}
		fmt.Fprintln(ev.Output, ev.formatValue(val))
		return nil, nil

	case *VarAssignNode:
		val, err := ev.evaluateExpr(s.Expr, env)
		if err != nil {
			return nil, err
		}
		env.Assign(s.Name, val)
		return val, nil

	case *MemberAssignNode:
		objVal, err := ev.evaluateExpr(s.Obj, env)
		if err != nil {
			return nil, err
		}
		exprVal, err := ev.evaluateExpr(s.Expr, env)
		if err != nil {
			return nil, err
		}

		if m, ok := objVal.(map[string]any); ok {
			m[s.Member] = exprVal
			return exprVal, nil
		}
		return nil, fmt.Errorf("obyek ora ndhukung properti %q", s.Member)

	case *IndexAssignNode:
		objVal, err := ev.evaluateExpr(s.Obj, env)
		if err != nil {
			return nil, err
		}
		idxVal, err := ev.evaluateExpr(s.Index, env)
		if err != nil {
			return nil, err
		}
		exprVal, err := ev.evaluateExpr(s.Expr, env)
		if err != nil {
			return nil, err
		}

		switch container := objVal.(type) {
		case map[string]any:
			keyStr := fmt.Sprintf("%v", idxVal)
			container[keyStr] = exprVal
			return exprVal, nil
		case []any:
			idx, ok := toInt(idxVal)
			if !ok || idx < 0 || idx >= len(container) {
				return nil, fmt.Errorf("indeks metu saka wates: %v", idxVal)
			}
			container[idx] = exprVal
			return exprVal, nil
		default:
			return nil, fmt.Errorf("tipe %T ora bisa diwenehi indeks", objVal)
		}

	case *IfNode:
		condVal, err := ev.evaluateExpr(s.Condition, env)
		if err != nil {
			return nil, err
		}

		if isTruthy(condVal) {
			blockEnv := NewEnvironment(env)
			return ev.executeBlock(s.TrueBranch, blockEnv)
		} else if len(s.FalseBranch) > 0 {
			blockEnv := NewEnvironment(env)
			return ev.executeBlock(s.FalseBranch, blockEnv)
		}
		return nil, nil

	case *ForNode:
		startVal, err := ev.evaluateExpr(s.StartExpr, env)
		if err != nil {
			return nil, err
		}
		endVal, err := ev.evaluateExpr(s.EndExpr, env)
		if err != nil {
			return nil, err
		}

		start, ok1 := toInt(startVal)
		end, ok2 := toInt(endVal)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("nilai baleni saka-nganti kudu angka")
		}

		step := 1
		if start > end {
			step = -1
		}

		for i := start; (step > 0 && i <= end) || (step < 0 && i >= end); i += step {
			loopEnv := NewEnvironment(env)
			loopEnv.Set(s.VarName, int64(i))
			res, err := ev.executeBlock(s.Body, loopEnv)
			if err != nil {
				return nil, err
			}
			if ret, ok := res.(returnSignal); ok {
				return ret, nil
			}
			if _, ok := res.(breakSignal); ok {
				break
			}
		}
		return nil, nil

	case *ForEachNode:
		collVal, err := ev.evaluateExpr(s.Collection, env)
		if err != nil {
			return nil, err
		}

		var items []any
		switch c := collVal.(type) {
		case []any:
			items = c
		case map[string]any:
			for k := range c {
				items = append(items, k)
			}
		case string:
			for _, r := range c {
				items = append(items, string(r))
			}
		default:
			return nil, fmt.Errorf("tipe %T ora bisa dibaleni nganggo 'neng'", collVal)
		}

		for _, item := range items {
			loopEnv := NewEnvironment(env)
			loopEnv.Set(s.VarName, item)
			res, err := ev.executeBlock(s.Body, loopEnv)
			if err != nil {
				return nil, err
			}
			if ret, ok := res.(returnSignal); ok {
				return ret, nil
			}
			if _, ok := res.(breakSignal); ok {
				break
			}
		}
		return nil, nil

	case *FuncDefNode:
		fn := &Function{
			Name:    s.Name,
			Params:  s.Params,
			Body:    s.Body,
			Closure: env,
		}
		if s.Name != "anonymous" {
			env.Set(s.Name, fn)
		}
		return fn, nil

	case *ReturnNode:
		var val any
		if s.Expr != nil {
			var err error
			val, err = ev.evaluateExpr(s.Expr, env)
			if err != nil {
				return nil, err
			}
		}
		return returnSignal{value: val}, nil

	case *BreakNode:
		return breakSignal{}, nil

	case *ContinueNode:
		return continueSignal{}, nil

	case *ExprStatement:
		return ev.evaluateExpr(s.Expr, env)

	default:
		return nil, fmt.Errorf("pranyatan ora dimangerteni: %T", stmt)
	}
}

func (ev *Evaluator) evaluateExpr(expr Expression, env *Environment) (any, error) {
	switch e := expr.(type) {
	case *LiteralNode:
		return e.Value, nil

	case *VarAccessNode:
		val, ok := env.Get(e.Name)
		if !ok {
			return nil, fmt.Errorf("variabel %q durung ditepungake", e.Name)
		}
		return val, nil

	case *BinOpNode:
		left, err := ev.evaluateExpr(e.Left, env)
		if err != nil {
			return nil, err
		}

		// Short-circuit logical operators
		if e.Op == T_LAN {
			if !isTruthy(left) {
				return false, nil
			}
			right, err := ev.evaluateExpr(e.Right, env)
			if err != nil {
				return nil, err
			}
			return isTruthy(right), nil
		}
		if e.Op == T_UTAWA {
			if isTruthy(left) {
				return true, nil
			}
			right, err := ev.evaluateExpr(e.Right, env)
			if err != nil {
				return nil, err
			}
			return isTruthy(right), nil
		}

		right, err := ev.evaluateExpr(e.Right, env)
		if err != nil {
			return nil, err
		}

		switch e.Op {
		case T_PLUS:
			if lStr, ok1 := left.(string); ok1 {
				return lStr + ev.formatValue(right), nil
			}
			if rStr, ok2 := right.(string); ok2 {
				return ev.formatValue(left) + rStr, nil
			}
			if lNum, ok1 := toFloat(left); ok1 {
				if rNum, ok2 := toFloat(right); ok2 {
					if isBothInt(left, right) {
						return int64(lNum) + int64(rNum), nil
					}
					return lNum + rNum, nil
				}
			}
			return nil, fmt.Errorf("operasi '+' ora ndhukung tipe %T lan %T", left, right)

		case T_MINUS:
			lNum, ok1 := toFloat(left)
			rNum, ok2 := toFloat(right)
			if ok1 && ok2 {
				if isBothInt(left, right) {
					return int64(lNum) - int64(rNum), nil
				}
				return lNum - rNum, nil
			}
			return nil, fmt.Errorf("operasi '-' mung kanggo angka")

		case T_MUL:
			if lStr, ok1 := left.(string); ok1 {
				if count, ok2 := toInt(right); ok2 && count >= 0 {
					return strings.Repeat(lStr, count), nil
				}
			}
			lNum, ok1 := toFloat(left)
			rNum, ok2 := toFloat(right)
			if ok1 && ok2 {
				if isBothInt(left, right) {
					return int64(lNum) * int64(rNum), nil
				}
				return lNum * rNum, nil
			}
			return nil, fmt.Errorf("operasi '*' mung kanggo angka utawa teks*angka")

		case T_DIV:
			lNum, ok1 := toFloat(left)
			rNum, ok2 := toFloat(right)
			if ok1 && ok2 {
				if rNum == 0 {
					return nil, fmt.Errorf("ora bisa mbagi nganggo nol")
				}
				if isBothInt(left, right) && int64(lNum)%int64(rNum) == 0 {
					return int64(lNum) / int64(rNum), nil
				}
				return lNum / rNum, nil
			}
			return nil, fmt.Errorf("operasi '/' mung kanggo angka")

		case T_EQ:
			return reflect.DeepEqual(left, right), nil
		case T_NEQ:
			return !reflect.DeepEqual(left, right), nil

		case T_LT:
			lNum, ok1 := toFloat(left)
			rNum, ok2 := toFloat(right)
			if ok1 && ok2 {
				return lNum < rNum, nil
			}
			if lStr, ok1 := left.(string); ok1 {
				if rStr, ok2 := right.(string); ok2 {
					return lStr < rStr, nil
				}
			}
			return false, fmt.Errorf("operasi '<' ora sah kanggo %T lan %T", left, right)

		case T_LTE:
			lNum, ok1 := toFloat(left)
			rNum, ok2 := toFloat(right)
			if ok1 && ok2 {
				return lNum <= rNum, nil
			}
			return false, nil

		case T_GT:
			lNum, ok1 := toFloat(left)
			rNum, ok2 := toFloat(right)
			if ok1 && ok2 {
				return lNum > rNum, nil
			}
			if lStr, ok1 := left.(string); ok1 {
				if rStr, ok2 := right.(string); ok2 {
					return lStr > rStr, nil
				}
			}
			return false, nil

		case T_GTE:
			lNum, ok1 := toFloat(left)
			rNum, ok2 := toFloat(right)
			if ok1 && ok2 {
				return lNum >= rNum, nil
			}
			return false, nil
		}

	case *UnaryOpNode:
		val, err := ev.evaluateExpr(e.Expr, env)
		if err != nil {
			return nil, err
		}
		if e.Op == T_ORA {
			return !isTruthy(val), nil
		}
		if e.Op == T_MINUS {
			if n, ok := toFloat(val); ok {
				if _, isInt := val.(int64); isInt {
					return -int64(n), nil
				}
				return -n, nil
			}
			return nil, fmt.Errorf("operasi '-' ora sah kanggo %T", val)
		}

	case *FuncDefNode:
		fn := &Function{
			Name:    e.Name,
			Params:  e.Params,
			Body:    e.Body,
			Closure: env,
		}
		return fn, nil

	case *FuncCallNode:
		calleeVal, ok := env.Get(e.Callee)
		if !ok {
			return nil, fmt.Errorf("fungsi %q ora ditemokake", e.Callee)
		}

		var args []any
		for _, argExpr := range e.Args {
			argVal, err := ev.evaluateExpr(argExpr, env)
			if err != nil {
				return nil, err
			}
			args = append(args, argVal)
		}

		return ev.CallCallable(calleeVal, args)

	case *MemberAccessNode:
		objVal, err := ev.evaluateExpr(e.Obj, env)
		if err != nil {
			return nil, err
		}
		if m, ok := objVal.(map[string]any); ok {
			return m[e.Member], nil
		}
		// Reflection fallback for Go structs/interfaces
		return ev.reflectMember(objVal, e.Member)

	case *MethodCallNode:
		objVal, err := ev.evaluateExpr(e.Obj, env)
		if err != nil {
			return nil, err
		}

		var args []any
		for _, argExpr := range e.Args {
			argVal, err := ev.evaluateExpr(argExpr, env)
			if err != nil {
				return nil, err
			}
			args = append(args, argVal)
		}

		// 1. If obj is a map containing a function/method
		if m, ok := objVal.(map[string]any); ok {
			if fnVal, exists := m[e.Method]; exists {
				return ev.CallCallable(fnVal, args)
			}
		}

		// 2. Reflect method call on Go struct
		return ev.reflectMethod(objVal, e.Method, args)

	case *IndexAccessNode:
		objVal, err := ev.evaluateExpr(e.Obj, env)
		if err != nil {
			return nil, err
		}
		idxVal, err := ev.evaluateExpr(e.Index, env)
		if err != nil {
			return nil, err
		}

		switch container := objVal.(type) {
		case map[string]any:
			keyStr := fmt.Sprintf("%v", idxVal)
			return container[keyStr], nil
		case []any:
			idx, ok := toInt(idxVal)
			if !ok || idx < 0 || idx >= len(container) {
				return nil, fmt.Errorf("indeks metu saka wates: %v", idxVal)
			}
			return container[idx], nil
		case string:
			idx, ok := toInt(idxVal)
			if !ok || idx < 0 || idx >= len(container) {
				return nil, fmt.Errorf("indeks metu saka wates: %v", idxVal)
			}
			return string(container[idx]), nil
		default:
			return nil, fmt.Errorf("tipe %T ora ndhukung indeks", objVal)
		}

	case *DictNode:
		dict := make(map[string]any)
		for i := 0; i < len(e.Keys); i++ {
			kVal, err := ev.evaluateExpr(e.Keys[i], env)
			if err != nil {
				return nil, err
			}
			vVal, err := ev.evaluateExpr(e.Values[i], env)
			if err != nil {
				return nil, err
			}
			dict[fmt.Sprintf("%v", kVal)] = vVal
		}
		return dict, nil

	case *ListNode:
		var list []any
		for _, elem := range e.Elements {
			elemVal, err := ev.evaluateExpr(elem, env)
			if err != nil {
				return nil, err
			}
			list = append(list, elemVal)
		}
		return list, nil
	}

	return nil, fmt.Errorf("ekspresi ora dimangerteni: %T", expr)
}

// CallCallable executes a Ngapak Function or BuiltinFunc.
func (ev *Evaluator) CallCallable(callee any, args []any) (any, error) {
	switch fn := callee.(type) {
	case *Function:
		callEnv := NewEnvironment(fn.Closure)
		for i, param := range fn.Params {
			if i < len(args) {
				callEnv.Set(param, args[i])
			} else {
				callEnv.Set(param, nil)
			}
		}

		res, err := ev.executeBlock(fn.Body, callEnv)
		if err != nil {
			return nil, err
		}
		if ret, ok := res.(returnSignal); ok {
			return ret.value, nil
		}
		return nil, nil

	case *BuiltinFunc:
		return fn.Fn(args)

	case func(args []any) (any, error):
		return fn(args)

	default:
		return nil, fmt.Errorf("obyek %T ora bisa diundang minangka fungsi", callee)
	}
}

func (ev *Evaluator) reflectMember(obj any, member string) (any, error) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		field := val.FieldByName(strings.Title(member))
		if field.IsValid() {
			return field.Interface(), nil
		}
		field = val.FieldByName(member)
		if field.IsValid() {
			return field.Interface(), nil
		}
	}
	return nil, fmt.Errorf("properti %q ora ditemokake ing %T", member, obj)
}

func (ev *Evaluator) reflectMethod(obj any, method string, args []any) (any, error) {
	val := reflect.ValueOf(obj)
	m := val.MethodByName(strings.Title(method))
	if !m.IsValid() {
		m = val.MethodByName(method)
	}
	if !m.IsValid() {
		return nil, fmt.Errorf("metode %q ora ditemokake ing %T", method, obj)
	}

	mType := m.Type()
	numIn := mType.NumIn()
	callArgs := make([]reflect.Value, numIn)

	for i := 0; i < numIn; i++ {
		if i < len(args) {
			callArgs[i] = reflect.ValueOf(args[i])
		} else {
			callArgs[i] = reflect.Zero(mType.In(i))
		}
	}

	results := m.Call(callArgs)
	if len(results) == 0 {
		return nil, nil
	}
	if len(results) == 1 {
		return results[0].Interface(), nil
	}
	// Return first value, and error if last value is error
	if err, ok := results[len(results)-1].Interface().(error); ok && err != nil {
		return nil, err
	}
	return results[0].Interface(), nil
}

func (ev *Evaluator) formatValue(v any) string {
	if v == nil {
		return "kosong"
	}
	if b, ok := v.(bool); ok {
		if b {
			return "bener"
		}
		return "salah"
	}
	return fmt.Sprintf("%v", v)
}

func isTruthy(v any) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	if n, ok := toFloat(v); ok {
		return n != 0
	}
	if s, ok := v.(string); ok {
		return len(s) > 0
	}
	return true
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case int64:
		return float64(n), true
	case int:
		return float64(n), true
	case float64:
		return n, true
	case float32:
		return float64(n), true
	default:
		return 0, false
	}
}

func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int64:
		return int(n), true
	case int:
		return n, true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}

func isBothInt(a, b any) bool {
	_, ok1 := a.(int64)
	if !ok1 {
		_, ok1 = a.(int)
	}
	_, ok2 := b.(int64)
	if !ok2 {
		_, ok2 = b.(int)
	}
	return ok1 && ok2
}

