package builtin

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

type xmlNode struct {
	Name     string
	Attrs    map[string]string
	Text     string
	Children []*xmlNode
}

func newXMLModule() Object {
	parseFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("xml.parse expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("xml.parse", args, 0)
		if err != nil {
			return nil, err
		}
		node, err := parseXMLNode(text)
		if err != nil {
			return nil, err
		}
		return xmlNodeToObject(node), nil
	})

	fromFileFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("xml.fromFile expects 1 arg, got %d", len(args))
		}
		path, err := asStringArg("xml.fromFile", args, 0)
		if err != nil {
			return nil, err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("xml.fromFile failed: %w", err)
		}
		node, err := parseXMLNode(string(data))
		if err != nil {
			return nil, fmt.Errorf("xml.fromFile parse error: %w", err)
		}
		return xmlNodeToObject(node), nil
	})

	stringifyFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("xml.stringify expects 1-2 args: node, [pretty]")
		}
		nodeObj, ok := args[0].(Object)
		if !ok {
			return nil, fmt.Errorf("xml.stringify arg[0] expects node object, got %T", args[0])
		}
		node, err := objectToXMLNode(nodeObj)
		if err != nil {
			return nil, err
		}
		pretty := false
		if len(args) == 2 {
			b, ok := args[1].(bool)
			if !ok {
				return nil, fmt.Errorf("xml.stringify arg[1] expects bool, got %T", args[1])
			}
			pretty = b
		}
		return stringifyXMLNode(node, pretty), nil
	})

	validFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("xml.valid expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("xml.valid", args, 0)
		if err != nil {
			return nil, err
		}
		_, err = parseXMLNode(text)
		return err == nil, nil
	})

	escapeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("xml.escape expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("xml.escape", args, 0)
		if err != nil {
			return nil, err
		}
		return escapeXMLText(text), nil
	})

	saveToFileFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 || len(args) > 3 {
			return nil, fmt.Errorf("xml.saveToFile expects 2-3 args: node, path, [pretty]")
		}
		nodeObj, ok := args[0].(Object)
		if !ok {
			return nil, fmt.Errorf("xml.saveToFile arg[0] expects node object, got %T", args[0])
		}
		path, err := asStringArg("xml.saveToFile", args, 1)
		if err != nil {
			return nil, err
		}
		pretty := false
		if len(args) == 3 {
			b, ok := args[2].(bool)
			if !ok {
				return nil, fmt.Errorf("xml.saveToFile arg[2] expects bool, got %T", args[2])
			}
			pretty = b
		}
		node, err := objectToXMLNode(nodeObj)
		if err != nil {
			return nil, err
		}
		content := stringifyXMLNode(node, pretty)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("xml.saveToFile write error: %w", err)
		}
		return true, nil
	})

	namespace := Object{
		"parse":      parseFn,
		"fromFile":   fromFileFn,
		"stringify":  stringifyFn,
		"saveToFile": saveToFileFn,
		"valid":      validFn,
		"escape":     escapeFn,
	}
	module := cloneObject(namespace)
	module["xml"] = namespace
	return module
}

func parseXMLNode(text string) (*xmlNode, error) {
	decoder := xml.NewDecoder(strings.NewReader(text))
	var stack []*xmlNode
	var root *xmlNode
	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			node := &xmlNode{
				Name:  t.Name.Local,
				Attrs: map[string]string{},
			}
			for _, attr := range t.Attr {
				node.Attrs[attr.Name.Local] = attr.Value
			}
			if len(stack) == 0 {
				root = node
			} else {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			}
			stack = append(stack, node)
		case xml.CharData:
			if len(stack) == 0 {
				continue
			}
			s := strings.TrimSpace(string(t))
			if s == "" {
				continue
			}
			current := stack[len(stack)-1]
			if current.Text == "" {
				current.Text = s
			} else {
				current.Text += s
			}
		case xml.EndElement:
			if len(stack) == 0 {
				return nil, fmt.Errorf("xml parse error: unexpected closing tag")
			}
			stack = stack[:len(stack)-1]
		}
	}
	if root == nil {
		return nil, fmt.Errorf("xml parse error: empty document")
	}
	if len(stack) != 0 {
		return nil, fmt.Errorf("xml parse error: unclosed tags")
	}
	return root, nil
}

func xmlNodeToObject(node *xmlNode) Object {
	attrs := Object{}
	for k, v := range node.Attrs {
		attrs[k] = v
	}
	children := make(Array, 0, len(node.Children))
	for _, child := range node.Children {
		children = append(children, xmlNodeToObject(child))
	}
	return Object{
		"name":     node.Name,
		"attrs":    attrs,
		"text":     node.Text,
		"children": children,
	}
}

func objectToXMLNode(obj Object) (*xmlNode, error) {
	nameRaw, ok := obj["name"]
	if !ok {
		return nil, fmt.Errorf("xml node requires name")
	}
	name, err := asStringValue("xml node.name", nameRaw)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("xml node.name must not be empty")
	}
	node := &xmlNode{
		Name:  name,
		Attrs: map[string]string{},
	}
	if attrsRaw, ok := obj["attrs"]; ok && attrsRaw != nil {
		attrsObj, ok := attrsRaw.(Object)
		if !ok {
			return nil, fmt.Errorf("xml node.attrs expects object, got %T", attrsRaw)
		}
		for k, v := range attrsObj {
			s, err := asStringValue("xml node.attrs["+k+"]", v)
			if err != nil {
				return nil, err
			}
			node.Attrs[k] = s
		}
	}
	if textRaw, ok := obj["text"]; ok && textRaw != nil {
		text, err := asStringValue("xml node.text", textRaw)
		if err != nil {
			return nil, err
		}
		node.Text = text
	}
	if childrenRaw, ok := obj["children"]; ok && childrenRaw != nil {
		childrenArr, ok := childrenRaw.(Array)
		if !ok {
			return nil, fmt.Errorf("xml node.children expects array, got %T", childrenRaw)
		}
		for i, c := range childrenArr {
			childObj, ok := c.(Object)
			if !ok {
				return nil, fmt.Errorf("xml node.children[%d] expects object, got %T", i, c)
			}
			child, err := objectToXMLNode(childObj)
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, child)
		}
	}
	return node, nil
}

func stringifyXMLNode(node *xmlNode, pretty bool) string {
	var b bytes.Buffer
	writeXMLNode(&b, node, pretty, 0)
	return b.String()
}

func writeXMLNode(b *bytes.Buffer, node *xmlNode, pretty bool, depth int) {
	indent := ""
	if pretty {
		indent = strings.Repeat("  ", depth)
		b.WriteString(indent)
	}
	b.WriteByte('<')
	b.WriteString(node.Name)
	for k, v := range node.Attrs {
		b.WriteByte(' ')
		b.WriteString(k)
		b.WriteString(`="`)
		b.WriteString(escapeXMLText(v))
		b.WriteByte('"')
	}
	if node.Text == "" && len(node.Children) == 0 {
		b.WriteString("/>")
		if pretty {
			b.WriteByte('\n')
		}
		return
	}
	b.WriteByte('>')
	if node.Text != "" {
		b.WriteString(escapeXMLText(node.Text))
	}
	if len(node.Children) > 0 {
		if pretty {
			b.WriteByte('\n')
		}
		for _, child := range node.Children {
			writeXMLNode(b, child, pretty, depth+1)
		}
		if pretty {
			b.WriteString(indent)
		}
	}
	b.WriteString("</")
	b.WriteString(node.Name)
	b.WriteByte('>')
	if pretty {
		b.WriteByte('\n')
	}
}

func escapeXMLText(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
