package macro

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/OutboundSpade/markf/logger"
	"github.com/OutboundSpade/markf/macro/exec"
)

var UnsafeMacros = []*Macro{
	{
		MacroName: "exec",
		MacroFunc: func(args *[]string) string {
			if len(*args) <= 1 {
				return "Usage: exec \\<command>"
			}
			res := exec.RunCommand(strings.Join((*args)[1:], " "))
			logger.Printf("exec: '%s'", *res)
			*res = bytes.TrimSpace(*res)
			return string(*res)
		},
	},
	{
		MacroName: "exec-screenshot",
		MacroFunc: func(args *[]string) string {
			if len(*args) <= 1 {
				return "Usage: exec-screenshot \\<command>"
			}
			res, err := exec.ScreenShotCommands((*args)[1:])
			if err != nil {
				return err.Error()
			}
			data := base64.StdEncoding.EncodeToString(*res)
			logger.Printf("exec-screenshot: '%s'", (*args)[1])
			return fmt.Sprintf("![](data:image/png;base64,%s)", data)
		},
	},
	{
		MacroName: "file-read",
		MacroFunc: func(args *[]string) string {
			if len(*args) <= 1 {
				return "Usage: file-read \\<filename>"
			}
			res, err := os.ReadFile((*args)[1])
			if err != nil {
				return err.Error()
			}
			return string(res)
		},
	},
}
