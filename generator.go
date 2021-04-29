package main

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

// 因为我们针对数据模型生成方法，所以要屏蔽掉的除此之外的 Message， 根名称为
// "Response", "Request" 的 Message 要被忽略掉。
var IgnoreSuffixes = []string{
	"Response",
	"Request",
}

type Generator struct {
	*generator.Generator
	generator.PluginImports
	write bool
}

func NewGenerator() *Generator {
	return &Generator{}
}

func (p *Generator) Name() string {
	return "sql"
}

func (p *Generator) Init(g *generator.Generator) {
	p.Generator = g
}

func (p *Generator) GenerateImports(file *generator.FileDescriptor) {
	if len(file.Messages()) == 0 {
		return
	}
	msgs := p.msgs(file)
	if len(msgs.Protos) > 0 {
		p.PrintImport("p", "google.golang.org/protobuf/proto")
		p.PrintImport("driver", "database/sql/driver")
		p.PrintImport("gorm", "gorm.io/gorm")
		p.PrintImport("schema", "gorm.io/gorm/schema")
	}

}

func (p *Generator) Generate(file *generator.FileDescriptor) {
	p.write = false
	t := template.Must(template.New("sql").Parse(tmpl))
	var buf bytes.Buffer
	t.Execute(&buf, p.msgs(file))
	p.P(buf.String())
}

func (p *Generator) Write() bool {
	return p.write
}

func forEachMessage(parent *descriptor.DescriptorProto, children []*descriptor.DescriptorProto, f func(parent *descriptor.DescriptorProto, child *descriptor.DescriptorProto)) {
	for _, child := range children {
		f(parent, child)
		forEachMessage(child, child.NestedType, f)
	}
}

func (p *Generator) msgs(file *generator.FileDescriptor) Msgs {
	var msgs Msgs

	forEachMessage(nil, file.MessageType, func(parent *descriptor.DescriptorProto, child *descriptor.DescriptorProto) {
		var name string
		if parent != nil {
			parentName := generator.CamelCase(*parent.Name)
			childName := generator.CamelCase(*child.Name)
			name = fmt.Sprintf("%s_%s", parentName, childName)
		} else {
			name = generator.CamelCase(*child.Name)
		}
		if child.Options != nil && child.Options.MapEntry != nil && *child.Options.MapEntry {
			return
		}
		child.Name = &name
		msgs.Protos = append(msgs.Protos, child)
	})

	protos := []*descriptor.DescriptorProto{}
	for _, m := range msgs.Protos {
		firstName := strings.Split(*m.Name, "_")[0]
		ignore := false
		for _, s := range IgnoreSuffixes {
			if strings.HasSuffix(firstName, s) {
				ignore = true
				break
			}
		}
		if !ignore {
			protos = append(protos, m)
		}
	}
	if len(protos) > 0 {
		p.write = true
	}
	msgs.Protos = protos
	return msgs
}

func init() {
	generator.RegisterPlugin(NewGenerator())
}

type Msgs struct {
	Protos []*descriptor.DescriptorProto
}

var tmpl = `
{{ range $message := .Protos }}
func (t *{{ $message.Name }}) Scan(val interface{}) error {
	return p.Unmarshal(val.([]byte), t)
}

func (t *{{ $message.Name }}) Value() (driver.Value, error) {
	if t == nil {
		t = &{{ $message.Name }}{}
	}
	return p.Marshal(t)
}

func (*{{ $message.Name }}) GormDBDataType(*gorm.DB, *schema.Field) string {
	return "blob"
}
{{ end }}
`
