package webapp

import (
	"embed"
	"path"
	"text/template"
)

//go:embed *
// Dir is webapp dir embed fs
var Dir embed.FS

// IndexHTML is index.html file name
var IndexHTML = "index.html"

// Webapp 返回 webapp 文件
func Webapp(name string, debug bool) (*template.Template, error) {
	if name == "" || name == "/" {
		name = IndexHTML
	} else {
		if extension := path.Ext(name); extension == "" {
			// 如果没有指定后缀，则直接返回 index.html 文件
			name = IndexHTML
		}
	}
	if debug {
		tmpl, err := template.ParseFiles("../webapp/" + name)
		if err != nil {
			tmpl, err = Webapp(IndexHTML, debug)
			if err != nil {
				return nil, err
			}
		}
		return tmpl, nil
	}

	tmpl := template.New(name)
	bs, err := Dir.ReadFile(name)
	if err != nil {
		tmpl, err = Webapp(IndexHTML, debug)
		if err != nil {
			return nil, err
		}
	}
	if _, err = tmpl.Parse(string(bs)); err != nil {
		tmpl, err = Webapp(IndexHTML, debug)
		if err != nil {
			return nil, err
		}
	}
	return tmpl, nil
}
