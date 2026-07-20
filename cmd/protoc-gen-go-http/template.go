package main

import (
	"bytes"
	"strings"
	"text/template"
)

var serviceTemplate = `
{{$physicalAddr := .PhysicalAddr}}
type {{.ServiceType}}HTTPServer interface {
{{- range .Methods}}
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{- end}}
}

func Register{{.ServiceType}}HTTPServer(s *http.Server, srv {{.ServiceType}}HTTPServer) {
	r := s.Route("/")
	{{- range .Methods}}
	r.{{.Method}}("{{.Path}}", _{{.ServiceType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv))
	{{- end}}
}

{{range .Methods}}
func _{{.ServiceType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv {{.ServiceType}}HTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in {{.Request}}
		http.SetOperation(ctx, Operation{{.ServiceType}}{{.Name}})
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			{{- if .RequestQuery}}
			if err := ctx.Bind(&req.(*{{.Request}}){{.RequestQuery}}); err != nil {
				return nil, err
			}
			{{- else}}
			if err := ctx.Bind(req); err != nil {
				return nil, err
			}
			{{- end}}
			if err := ctx.BindQuery(req); err != nil {
				return nil, err
			}
			return srv.{{.Name}}(ctx, req.(*{{.Request}}))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*{{.Reply}})
		return ctx.Result(200, reply)
	}
}
{{end}}
`

func (s *service) execute() string {
	if s.pathTemplate == "" {
		s.pathTemplate = serviceTemplate
	}
	buf := new(bytes.Buffer)
	tmpl, err := template.New("http").Parse(strings.TrimSpace(s.pathTemplate))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return buf.String()
}
