package main

import (
	"bytes"
	"fmt"
	"html/template"
)

func generateSSZSize(typ SSZType) ([]byte, error) {
	tmpl, err := template.New("ssz-size").Parse(`func ({{.Receiver}} {{.Type}}) SizeSSZ() int {
	size := {{.FixedSize}}
	{{- .VariableSize }}
	return size
}`)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(nil)

	err = tmpl.Execute(buf, struct {
		Receiver     string
		Type         string
		FixedSize    int
		VariableSize string
	}{
		Receiver:  typ.Name(),
		Type:      fmt.Sprintf("*%s", typ.Name()),
		FixedSize: typ.FixedSize(),
		//VariableSize: "\n" + strings.Join(variableComputations, "\n"),
	})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func generate(typ SSZType) ([]byte, error) {
	size, err := generateSSZSize(typ)
	if err != nil {
		return nil, err
	}
	return size, nil
}
