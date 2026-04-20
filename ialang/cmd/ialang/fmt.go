package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"ialang/pkg/lang"
	astpkg "ialang/pkg/lang/ast"
	tok "ialang/pkg/lang/token"
)

// Operator precedence levels (must match parser.go)
const (
	_ int = iota
	LOWEST
	TERTIARY
	LOGICALOR
	LOGICALAND
	EQUALS
	COMPARE
	SUM
	PRODUCT
	PREFIX
	CALL
	MEMBER
	INDEX
)

var operatorPrecedence = map[tok.TokenType]int{
	tok.OR:       LOGICALOR,
	tok.AND:      LOGICALAND,
	tok.EQ:       EQUALS,
	tok.NEQ:      EQUALS,
	tok.LT:       COMPARE,
	tok.GT:       COMPARE,
	tok.LTE:      COMPARE,
	tok.GTE:      COMPARE,
	tok.PLUS:     SUM,
	tok.MINUS:    SUM,
	tok.BITOR:    SUM,
	tok.BITXOR:   SUM,
	tok.BITAND:   SUM,
	tok.SHL:      PRODUCT,
	tok.SHR:      PRODUCT,
	tok.ASTERISK: PRODUCT,
	tok.SLASH:    PRODUCT,
	tok.MODULO:   PRODUCT,
}

func exprPrecedence(expr astpkg.Expression) int {
	switch e := expr.(type) {
	case *astpkg.BinaryExpression:
		if p, ok := operatorPrecedence[e.Operator]; ok {
			return p
		}
		return LOWEST
	case *astpkg.UnaryExpression:
		return PREFIX
	case *astpkg.CallExpression, *astpkg.NewExpression:
		return CALL
	case *astpkg.GetExpression, *astpkg.IndexExpression:
		return MEMBER
	case *astpkg.TernaryExpression:
		return TERTIARY
	default:
		return LOWEST
	}
}

func needsParentheses(child, parent astpkg.Expression, isRightOperand bool) bool {
	childPrec := exprPrecedence(child)
	parentPrec := exprPrecedence(parent)
	if childPrec > parentPrec {
		return false
	}
	if childPrec < parentPrec {
		return true
	}
	// Same precedence: check associativity
	// Most operators are left-associative, so right operand needs parens at same level
	if isRightOperand && childPrec == parentPrec {
		return true
	}
	return false
}

func executeFmtCommand(path string, stdout, stderr io.Writer) error {
	target := strings.TrimSpace(path)
	if target == "" {
		return fmt.Errorf("fmt expects a file path")
	}

	// Check if target is a directory
	info, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("path not found: %s", target)
	}

	if info.IsDir() {
		return formatDirectory(target, stdout, stderr)
	}

	return formatFile(target, stdout, stderr)
}

func formatFile(path string, stdout, stderr io.Writer) error {
	src, err := readRunSource(path)
	if err != nil {
		return err
	}

	l := lang.NewLexer(src)
	p := lang.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		fmt.Fprintf(stderr, "parse errors in %s:\n", path)
		for _, e := range p.Errors() {
			fmt.Fprintf(stderr, "- %s\n", e)
		}
		return fmt.Errorf("parse failed")
	}

	formatted := formatProgramSource(program)
	if !strings.HasSuffix(formatted, "\n") {
		formatted += "\n"
	}

	if normalizeNewlines(src) == formatted {
		fmt.Fprintf(stdout, "already formatted: %s\n", path)
		return nil
	}

	if err := os.WriteFile(path, []byte(formatted), 0o644); err != nil {
		return fmt.Errorf("write formatted file error: %w", err)
	}
	fmt.Fprintf(stdout, "formatted: %s\n", path)
	return nil
}

func formatDirectory(dir string, stdout, stderr io.Writer) error {
	var files []string
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".ia") {
			// Skip common directories that shouldn't be formatted
			skipDirs := []string{"node_modules", ".git", ".tmp", ".gocache", ".gomodcache", ".gopath"}
			for _, skipDir := range skipDirs {
				if strings.Contains(path, skipDir) {
					return nil
				}
			}
			files = append(files, path)
		}
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("error scanning directory: %w", err)
	}
	
	if len(files) == 0 {
		fmt.Fprintf(stdout, "no .ia files found in %s\n", dir)
		return nil
	}
	
	fmt.Fprintf(stdout, "found %d .ia file(s) in %s\n\n", len(files), dir)
	
	formattedCount := 0
	errorCount := 0
	
	for _, file := range files {
		err := formatFile(file, stdout, stderr)
		if err != nil {
			fmt.Fprintf(stderr, "error formatting %s: %v\n", file, err)
			errorCount++
		} else {
			formattedCount++
		}
	}
	
	fmt.Fprintf(stdout, "\nsummary: %d formatted, %d errors, %d total\n", formattedCount, errorCount, len(files))
	
	if errorCount > 0 {
		return fmt.Errorf("some files failed to format")
	}
	
	return nil
}

func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

func formatProgramSource(program *astpkg.Program) string {
	if program == nil {
		return ""
	}
	f := &sourceFormatter{}
	
	for i, stmt := range program.Statements {
		// Write leading comments
		for _, comment := range stmt.GetLeadingComments() {
			f.writeIndent()
			f.b.WriteString(comment.Text)
			f.b.WriteByte('\n')
		}
		
		f.writeStatement(stmt)
		
		// Add newline between statements
		if i < len(program.Statements)-1 {
			f.b.WriteByte('\n')
			// Add extra blank line between imports and non-imports, or before/after functions/classes
			if shouldAddBlankLine(stmt, program.Statements[i+1]) {
				f.b.WriteByte('\n')
			}
		}
	}
	return f.b.String()
}

func shouldAddBlankLine(current, next astpkg.Statement) bool {
	// Don't add blank line between consecutive imports
	if _, isImport := current.(*astpkg.ImportStatement); isImport {
		if _, nextIsImport := next.(*astpkg.ImportStatement); nextIsImport {
			return false
		}
		// Add blank line after last import
		return true
	}
	// Add blank line before first import (if any statement precedes it)
	if _, nextIsImport := next.(*astpkg.ImportStatement); nextIsImport {
		return true
	}
	// Add blank line before/after functions and classes (including exported ones)
	switch next.(type) {
	case *astpkg.FunctionStatement, *astpkg.ClassStatement:
		return true
	case *astpkg.ExportStatement:
		// Check if export contains function/class
		if exp, ok := next.(*astpkg.ExportStatement); ok {
			switch exp.Statement.(type) {
			case *astpkg.FunctionStatement, *astpkg.ClassStatement:
				return true
			}
		}
	}
	switch current.(type) {
	case *astpkg.FunctionStatement, *astpkg.ClassStatement:
		return true
	case *astpkg.ExportStatement:
		// Check if export contains function/class
		if exp, ok := current.(*astpkg.ExportStatement); ok {
			switch exp.Statement.(type) {
			case *astpkg.FunctionStatement, *astpkg.ClassStatement:
				return true
			}
		}
	}
	return false
}

type sourceFormatter struct {
	b      strings.Builder
	indent int
}

func (f *sourceFormatter) writeIndent() {
	for i := 0; i < f.indent; i++ {
		f.b.WriteString("  ")
	}
}

func (f *sourceFormatter) writeStatement(stmt astpkg.Statement) {
	switch s := stmt.(type) {
	case *astpkg.ImportStatement:
		f.b.WriteString("import { ")
		f.b.WriteString(strings.Join(s.Names, ", "))
		f.b.WriteString(" } from ")
		f.b.WriteString(quoteIAString(s.Module))
		f.b.WriteByte(';')
	case *astpkg.ExportStatement:
		f.b.WriteString("export ")
		f.writeExportTarget(s.Statement)
	case *astpkg.LetStatement:
		f.b.WriteString("let ")
		f.b.WriteString(s.Name)
		f.b.WriteString(" = ")
		f.b.WriteString(f.formatExpression(s.Initializer))
		f.b.WriteByte(';')
	case *astpkg.AssignStatement:
		f.b.WriteString(s.Name)
		f.b.WriteString(" = ")
		f.b.WriteString(f.formatExpression(s.Value))
		f.b.WriteByte(';')
	case *astpkg.CompoundAssignStatement:
		f.b.WriteString(s.Name)
		f.b.WriteByte(' ')
		f.b.WriteString(string(s.Operator))
		f.b.WriteByte(' ')
		f.b.WriteString(f.formatExpression(s.Value))
		f.b.WriteByte(';')
	case *astpkg.SetPropertyStatement:
		f.b.WriteString(f.formatPropertyTarget(s.Object, s.Property))
		f.b.WriteString(" = ")
		f.b.WriteString(f.formatExpression(s.Value))
		f.b.WriteByte(';')
	case *astpkg.CompoundSetPropertyStatement:
		f.b.WriteString(f.formatPropertyTarget(s.Object, s.Property))
		f.b.WriteByte(' ')
		f.b.WriteString(string(s.Operator))
		f.b.WriteByte(' ')
		f.b.WriteString(f.formatExpression(s.Value))
		f.b.WriteByte(';')
	case *astpkg.ExpressionStatement:
		f.b.WriteString(f.formatExpression(s.Expr))
		f.b.WriteByte(';')
	case *astpkg.FunctionStatement:
		if s.Async {
			f.b.WriteString("async ")
		}
		f.b.WriteString("function ")
		f.b.WriteString(s.Name)
		f.b.WriteByte('(')
		f.b.WriteString(strings.Join(s.Params, ", "))
		f.b.WriteString(") ")
		f.writeBlock(s.Body)
	case *astpkg.ClassStatement:
		f.b.WriteString("class ")
		f.b.WriteString(s.Name)
		if s.ParentName != "" {
			f.b.WriteString(" extends ")
			f.b.WriteString(s.ParentName)
		}
		if len(s.Methods) == 0 {
			f.b.WriteString(" {}")
			return
		}
		f.b.WriteString(" {\n")
		f.indent++
		for i, m := range s.Methods {
			f.writeIndent()
			if m.Async {
				f.b.WriteString("async ")
			}
			f.b.WriteString(m.Name)
			f.b.WriteByte('(')
			f.b.WriteString(strings.Join(m.Params, ", "))
			f.b.WriteString(") ")
			f.writeBlock(m.Body)
			if i < len(s.Methods)-1 {
				f.b.WriteByte('\n')
			}
		}
		f.indent--
		f.b.WriteByte('\n')
		f.writeIndent()
		f.b.WriteByte('}')
	case *astpkg.ReturnStatement:
		f.b.WriteString("return")
		if s.Value != nil {
			f.b.WriteByte(' ')
			f.b.WriteString(f.formatExpression(s.Value))
		}
		f.b.WriteByte(';')
	case *astpkg.ThrowStatement:
		f.b.WriteString("throw")
		if s.Value != nil {
			f.b.WriteByte(' ')
			f.b.WriteString(f.formatExpression(s.Value))
		}
		f.b.WriteByte(';')
	case *astpkg.BreakStatement:
		f.b.WriteString("break;")
	case *astpkg.ContinueStatement:
		f.b.WriteString("continue;")
	case *astpkg.TryCatchStatement:
		f.b.WriteString("try ")
		f.writeBlock(s.TryBlock)
		if s.CatchBlock != nil {
			f.b.WriteString(" catch (")
			f.b.WriteString(s.CatchName)
			f.b.WriteString(") ")
			f.writeBlock(s.CatchBlock)
		}
		if s.FinallyBlock != nil {
			f.b.WriteString(" finally ")
			f.writeBlock(s.FinallyBlock)
		}
	case *astpkg.IfStatement:
		f.b.WriteString("if (")
		f.b.WriteString(f.formatExpression(s.Condition))
		f.b.WriteString(") ")
		f.writeBlock(s.Then)
		if s.Else != nil {
			f.b.WriteString(" else ")
			f.writeBlock(s.Else)
		}
	case *astpkg.WhileStatement:
		f.b.WriteString("while (")
		f.b.WriteString(f.formatExpression(s.Condition))
		f.b.WriteString(") ")
		f.writeBlock(s.Body)
	case *astpkg.ForStatement:
		f.b.WriteString("for (")
		f.b.WriteString(f.formatForPart(s.Init))
		f.b.WriteString("; ")
		if s.Condition != nil {
			f.b.WriteString(f.formatExpression(s.Condition))
		}
		f.b.WriteString("; ")
		f.b.WriteString(f.formatForPart(s.Post))
		f.b.WriteString(") ")
		f.writeBlock(s.Body)
	default:
		f.b.WriteString("/* unsupported statement */")
	}
}

func (f *sourceFormatter) writeExportTarget(stmt astpkg.Statement) {
	switch s := stmt.(type) {
	case *astpkg.LetStatement:
		f.b.WriteString("let ")
		f.b.WriteString(s.Name)
		f.b.WriteString(" = ")
		f.b.WriteString(f.formatExpression(s.Initializer))
		f.b.WriteByte(';')
	case *astpkg.FunctionStatement:
		if s.Async {
			f.b.WriteString("async ")
		}
		f.b.WriteString("function ")
		f.b.WriteString(s.Name)
		f.b.WriteByte('(')
		f.b.WriteString(strings.Join(s.Params, ", "))
		f.b.WriteString(") ")
		f.writeBlock(s.Body)
	case *astpkg.ClassStatement:
		f.writeStatement(s)
	default:
		f.b.WriteString("/* unsupported export */")
	}
}

func (f *sourceFormatter) writeBlock(block *astpkg.BlockStatement) {
	if block == nil || len(block.Statements) == 0 {
		f.b.WriteString("{}")
		return
	}
	f.b.WriteString("{\n")
	f.indent++
	for i, stmt := range block.Statements {
		// Write leading comments
		for _, comment := range stmt.GetLeadingComments() {
			f.writeIndent()
			f.b.WriteString(comment.Text)
			f.b.WriteByte('\n')
		}
		
		f.writeIndent()
		f.writeStatement(stmt)
		if i < len(block.Statements)-1 {
			f.b.WriteByte('\n')
		}
	}
	f.indent--
	f.b.WriteByte('\n')
	f.writeIndent()
	f.b.WriteByte('}')
}

func (f *sourceFormatter) formatBlockStatement(block *astpkg.BlockStatement) string {
	var nested sourceFormatter
	nested.indent = f.indent
	nested.writeBlock(block)
	return nested.b.String()
}

func (f *sourceFormatter) formatForPart(stmt astpkg.Statement) string {
	switch s := stmt.(type) {
	case nil:
		return ""
	case *astpkg.LetStatement:
		return "let " + s.Name + " = " + f.formatExpression(s.Initializer)
	case *astpkg.AssignStatement:
		return s.Name + " = " + f.formatExpression(s.Value)
	case *astpkg.CompoundAssignStatement:
		return s.Name + " " + string(s.Operator) + " " + f.formatExpression(s.Value)
	case *astpkg.SetPropertyStatement:
		return f.formatPropertyTarget(s.Object, s.Property) + " = " + f.formatExpression(s.Value)
	case *astpkg.CompoundSetPropertyStatement:
		return f.formatPropertyTarget(s.Object, s.Property) + " " + string(s.Operator) + " " + f.formatExpression(s.Value)
	case *astpkg.ExpressionStatement:
		return f.formatExpression(s.Expr)
	default:
		return ""
	}
}

func (f *sourceFormatter) formatExpression(expr astpkg.Expression) string {
	switch e := expr.(type) {
	case nil:
		return ""
	case *astpkg.Identifier:
		return e.Name
	case *astpkg.NumberLiteral:
		return e.Value
	case *astpkg.StringLiteral:
		return quoteIAString(e.Value)
	case *astpkg.BoolLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *astpkg.NullLiteral:
		return "null"
	case *astpkg.ArrayLiteral:
		parts := make([]string, 0, len(e.Elements))
		for _, elem := range e.Elements {
			parts = append(parts, f.formatExpression(elem))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case *astpkg.ObjectLiteral:
		parts := make([]string, 0, len(e.Properties))
		for _, prop := range e.Properties {
			parts = append(parts, formatObjectKey(prop.Key)+": "+f.formatExpression(prop.Value))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case *astpkg.BinaryExpression:
		return f.formatBinaryExpression(e, LOWEST)
	case *astpkg.UnaryExpression:
		return string(e.Operator) + f.formatExpression(e.Right)
	case *astpkg.CallExpression:
		args := make([]string, 0, len(e.Arguments))
		for _, arg := range e.Arguments {
			args = append(args, f.formatExpression(arg))
		}
		return f.formatCallable(e.Callee) + "(" + strings.Join(args, ", ") + ")"
	case *astpkg.FunctionExpression:
		prefix := "function"
		if e.Async {
			prefix = "async function"
		}
		params := strings.Join(e.Params, ", ")
		name := ""
		if e.Name != "" {
			name = " " + e.Name
		}
		return prefix + name + "(" + params + ") " + f.formatBlockStatement(e.Body)
	case *astpkg.NewExpression:
		args := make([]string, 0, len(e.Arguments))
		for _, arg := range e.Arguments {
			args = append(args, f.formatExpression(arg))
		}
		return "new " + f.formatCallable(e.Callee) + "(" + strings.Join(args, ", ") + ")"
	case *astpkg.GetExpression:
		return f.formatMemberBase(e.Object) + "." + e.Property
	case *astpkg.IndexExpression:
		return f.formatMemberBase(e.Object) + "[" + f.formatExpression(e.Index) + "]"
	case *astpkg.AwaitExpression:
		return "await " + f.formatExpression(e.Expr)
	case *astpkg.TernaryExpression:
		return f.formatExpression(e.Condition) + " ? " + f.formatExpression(e.Then) + " : " + f.formatExpression(e.Else)
	case *astpkg.SuperExpression:
		return "super." + e.Property
	case *astpkg.SuperCallExpression:
		args := make([]string, 0, len(e.Arguments))
		for _, arg := range e.Arguments {
			args = append(args, f.formatExpression(arg))
		}
		return "super(" + strings.Join(args, ", ") + ")"
	default:
		return "undefined"
	}
}

func (f *sourceFormatter) formatBinaryExpression(e *astpkg.BinaryExpression, parentPrec int) string {
	leftPrec := exprPrecedence(e.Left)
	rightPrec := exprPrecedence(e.Right)
	currentPrec := exprPrecedence(e)

	var leftStr, rightStr string

	// Only add parentheses if child is a BinaryExpression with lower precedence
	if leftPrec < currentPrec && isBinaryLike(e.Left) {
		leftStr = "(" + f.formatExpression(e.Left) + ")"
	} else {
		leftStr = f.formatExpression(e.Left)
	}

	if rightPrec < currentPrec && isBinaryLike(e.Right) {
		rightStr = "(" + f.formatExpression(e.Right) + ")"
	} else {
		rightStr = f.formatExpression(e.Right)
	}

	result := leftStr + " " + string(e.Operator) + " " + rightStr

	if currentPrec < parentPrec {
		return "(" + result + ")"
	}
	return result
}

func isBinaryLike(expr astpkg.Expression) bool {
	switch expr.(type) {
	case *astpkg.BinaryExpression, *astpkg.TernaryExpression:
		return true
	default:
		return false
	}
}

func (f *sourceFormatter) formatPropertyTarget(object astpkg.Expression, property string) string {
	return f.formatMemberBase(object) + "." + property
}

func (f *sourceFormatter) formatMemberBase(expr astpkg.Expression) string {
	if isSimpleMemberBase(expr) {
		return f.formatExpression(expr)
	}
	return "(" + f.formatExpression(expr) + ")"
}

func (f *sourceFormatter) formatCallable(expr astpkg.Expression) string {
	if isSimpleCallable(expr) {
		return f.formatExpression(expr)
	}
	return "(" + f.formatExpression(expr) + ")"
}

func isSimpleMemberBase(expr astpkg.Expression) bool {
	switch expr.(type) {
	case *astpkg.Identifier, *astpkg.GetExpression, *astpkg.IndexExpression, *astpkg.CallExpression, *astpkg.NewExpression, *astpkg.SuperExpression, *astpkg.SuperCallExpression:
		return true
	default:
		return false
	}
}

func isSimpleCallable(expr astpkg.Expression) bool {
	switch expr.(type) {
	case *astpkg.Identifier, *astpkg.GetExpression, *astpkg.IndexExpression, *astpkg.CallExpression, *astpkg.SuperExpression, *astpkg.SuperCallExpression:
		return true
	default:
		return false
	}
}

func quoteIAString(s string) string {
	return `"` + s + `"`
}

func formatObjectKey(key string) string {
	if isIdentifierName(key) && tok.LookupIdent(key) == tok.IDENT {
		return key
	}
	return quoteIAString(key)
}

func isIdentifierName(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if i == 0 {
			if !isASCIILetter(ch) && ch != '_' {
				return false
			}
			continue
		}
		if !isASCIILetter(ch) && !isASCIIDigit(ch) && ch != '_' {
			return false
		}
	}
	return true
}

func isASCIILetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isASCIIDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
