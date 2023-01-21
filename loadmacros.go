package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/OutboundSpade/markf/logger"
	"github.com/OutboundSpade/markf/macro"
)

func loadMacros(p *macro.Parser) {
	if os.Getenv("MARKF_MACROS") != "" {
		loadMacrosFromDir(os.Getenv("MARKF_MACROS"), p)
	}
	cwd, _ := os.Getwd()
	localPath := filepath.Join(cwd, ".markf-macros")
	if _, err := os.Stat(localPath); !os.IsNotExist(err) {
		loadMacrosFromDir(".markf-macros", p)
	}
	hdir, _ := os.UserHomeDir()
	homePath := filepath.Join(hdir, ".markf-macros")
	if _, err := os.Stat(homePath); !os.IsNotExist(err) {
		loadMacrosFromDir(homePath, p)
	}

}

func loadMacrosFromDir(dir string, p *macro.Parser) {
	logger.Printf("Loading macros from %s\n", dir)
	files := []string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			files = append(files, path)
		}
		return nil
	},
	)
	must(err)
	for _, f := range files {
		loadMacro(f, p)
	}
}

var external_macros map[string]*[]byte = make(map[string]*[]byte)

func loadMacro(path string, p *macro.Parser) {
	logger.Printf("Loading macro from %s\n", path)
	f, err := os.Open(path)
	must(err)
	defer f.Close()
	data, err := io.ReadAll(f)
	must(err)
	name := filepath.Base(path)
	name = name[:len(name)-len(filepath.Ext(name))]
	external_macros[name] = &data
	logger.Printf("Registering macro %s\n", name)
	p.RegisterMacro(&macro.Macro{
		MacroName: name,
		MacroFunc: func(args *[]string) string {
			var newData []byte = make([]byte, len(*external_macros[name]))
			copy(newData, *external_macros[name])
			logger.Printf("newdata: %s\n", newData)
			var joined string
			if args == nil {
				args = &[]string{}
			}
			for i, arg := range *args {
				logger.Printf("arg %d: '%s'\n", i, arg)
				newData = bytes.ReplaceAll(newData, []byte(fmt.Sprintf("#$%d", i)), []byte(arg))
				if i == 0 {
					joined += fmt.Sprintf("`%s", arg)
				} else {
					joined += fmt.Sprintf("|%s`", arg)
				}
			}
			newData = bytes.ReplaceAll(newData, []byte("#$..."), []byte(joined))
			logger.Printf("args: '%s'\n", joined)
			return string(newData)
		},
		IsExternal: true,
	})
}
