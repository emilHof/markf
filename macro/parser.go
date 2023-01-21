package macro

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/OutboundSpade/markf/logger"
)

type OutputType uint

type Parser struct {
	macros []*Macro
}
type Macro struct {
	MacroName  string
	IsExternal bool
	MacroFunc  func(*[]string) string
}

var command_selector = regexp.MustCompile(`#!\(`)

// var arg_selector = regexp.MustCompile(`#\$((\d+)|...)`)
var paren_selector = regexp.MustCompile(`\(|\)`)
var curly_bracket_selector = regexp.MustCompile(`\{|\}`)

type paren struct {
	index  int
	isOpen bool
}

func (p Parser) Process(file *[]byte) {

	for i := command_selector.FindIndex(*file); i != nil; i = command_selector.FindIndex(*file) {
		offset := isolateFirstCommand((*file)[i[0]:]) //[]int{i[0], parenCollection[index-1].index + 1}
		if offset == nil {
			// line, col := getLineAndColNumbersFromIndex(file, i[0])
			// logger.Printf("%d\n", i[0])
			const debug_length = 20
			if len(*file) < i[0]+debug_length {
				logger.Printf("Unmatched parentheses near '%s'\n", (*file)[i[0]:])
			} else {
				logger.Printf("Unmatched parentheses near '%s'\n", (*file)[i[0]:i[0]+debug_length])
			}
			return
		}
		realIndecies := []int{i[0] + offset[0], i[0] + offset[1]}
		// logger.Printf("realIndecies: %v\n", realIndecies)
		// prevFoundIndexes = append(prevFoundIndexes, realIndecies)
		content := (*file)[realIndecies[0]:realIndecies[1]]
		// logger.Printf("evaluating macro: '%s'\n", content)
		res := EvalMacro(content, p.macros)
		if res == nil {
			res = new(string)
			*res = "Macro evaluation failed!"
			eval_count = 0
		}
		// logger.Printf("result: '%s'\n", *res)
		byteRes := []byte(*res)
		inlineReplaceBytes(file, []int{realIndecies[0], realIndecies[1]}, &byteRes)
		// logger.Printf("file: '%s'\n", *file)
	}
	*file = bytes.ReplaceAll(*file, []byte(`\n`), []byte("\n"))
	*file = bytes.ReplaceAll(*file, []byte(`\t`), []byte("\t"))
}

var quotes = regexp.MustCompile("[\"'`]")
var whitespace = regexp.MustCompile("\\s+")
var value = regexp.MustCompile("[^\"'`\\(\\)\\s]+")
var start_paren = regexp.MustCompile(`\(`)
var end_paren = regexp.MustCompile(`\)`)
var start_curly = regexp.MustCompile(`\{`)
var end_curly = regexp.MustCompile(`\}`)

var MAX_EVAL = 2048
var eval_count = 0

// var prevOut [][]byte

func EvalMacro(content []byte, macros []*Macro) *string {
	//strip prefix and suffix
	eval_count++
	if eval_count > MAX_EVAL {
		logger.Printf("Max eval count reached! Command: '%s'\n", content)
		return nil
	}
	content = bytes.TrimPrefix(content, []byte("#!("))
	content = bytes.TrimSuffix(content, []byte(")"))
	content = bytes.TrimSpace(content)
	logger.Printf("content: '%s'\n", content)
	s := bufio.NewScanner(bytes.NewReader(content))
	var args []string

	s.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF && len(data) == 0 {
			logger.Printf("EOF\n")
			return 0, nil, nil
		}
		//check for whitespace
		if i := whitespace.FindIndex(data); i != nil && i[0] == 0 {
			return i[1], []byte{}, nil
		}
		// if len(prevOut) > 0 {
		// 	offset := 0
		// 	if len(prevOut) == 1 {
		// 		logger.Printf("len(prevOut) == 1\n")
		// 		offset = 4
		// 		eval_count--
		// 	}
		// 	lastI := prevOut[0]
		// 	prevOut = prevOut[1:]
		// 	logger.Printf("lastI: '%s'\n", lastI)
		// 	// logger.Println("eval_depth: ", eval_count)
		// 	return len(lastI) + offset, lastI, nil
		// }
		logger.Printf("data: '%s'\n", data)
		//check if literal block
		if bytes.HasPrefix(data, []byte("{")) {
			logger.Printf("literal block\n")
			rIndecies := isolateCurlyBracketSet(data)
			if rIndecies != nil {
				formatted := bytes.TrimSpace(data[rIndecies[0]+1 : rIndecies[1]-1])
				formatted = append([]byte{'{'}, formatted...)
				formatted = append(formatted, []byte{'}'}...)
				logger.Printf("formatted: '%s'\n", formatted)
				return rIndecies[1] - rIndecies[0], formatted, nil
				// return rIndecies[1] - rIndecies[0], data[rIndecies[0]:rIndecies[1]], nil
			} else {
				return 0, nil, nil
			}
			// i := bytes.Index(data, []byte("}"))
			// tdata := append(append([]byte{'{'}, bytes.TrimSpace(data[1:i])...), '}')
			// tdata = bytes.ReplaceAll(tdata, []byte("\\n"), []byte{'\n'})
			// tdata = bytes.ReplaceAll(tdata, []byte("\\t"), []byte{'\t'})
			// tdata = bytes.ReplaceAll(tdata, []byte("\\s"), []byte{' '})
			// ti := bytes.Index(tdata, []byte("}"))
			// logger.Printf("data trimmed: '%s'\n", tdata)
			// if i >= 0 {
			// 	return i + 1, tdata[:ti+1], nil
			// }
			// return 0, nil, nil
		}
		//check if block
		if bytes.HasPrefix(data, []byte("#!(")) {
			rIndecies := isolateFirstCommand(data)
			if rIndecies != nil {
				return rIndecies[1] - rIndecies[0], data[rIndecies[0]:rIndecies[1]], nil
				// logger.Printf("rIndeciesMacro: '%v'\n", rIndecies)
				// mc := data[rIndecies[0]:rIndecies[1]]
				// logger.Printf("subcmd: '%s'\n", mc)
				// eval_count++
				// subargs := bytes.Split([]byte(*EvalMacro(mc, macros)), []byte(" "))
				// logger.Printf("subargs: '%s'\n", subargs)

				// prevOut = append(prevOut, subargs[1:]...)

				// return len(subargs[0]), subargs[0], nil
			} else {
				return 0, nil, nil
			}
		}

		// check if in paren
		if bytes.HasPrefix(data, []byte("(")) {
			rIndecies := isolateBracketSet(data)
			if rIndecies != nil {
				return rIndecies[1] - rIndecies[0], data[rIndecies[0]+1 : rIndecies[1]-1], nil
			} else {
				return 0, nil, nil
			}
			// parenC := len(start_paren.FindAllIndex(data, -1))
			// ends := end_paren.FindAllIndex(data, -1)
			// if len(ends) >= parenC {
			// 	return ends[parenC-1][1], data[:ends[parenC-1][1]], nil
			// } else {
			// 	return 0, nil, nil
			// }
		}

		//check for quotes
		if i := quotes.FindIndex(data); i != nil && i[0] == 0 {
			quote := data[i[0]]
			// logger.Printf("quote: '%s'\n", string(quote))
			//find end quote
			for j := 1; j < len(data); j++ {
				if data[j] == quote && data[j-1] != '\\' {
					return j + 1, data[:j+1], nil
				}
			}
			return 0, nil, nil
		}

		if i := value.FindIndex(data); i != nil && i[0] == 0 {
			return i[1], data[i[0]:i[1]], nil
		}

		// Request more data.
		return 0, nil, nil
	})
	for tok := s.Scan(); tok; tok = s.Scan() {
		tt := s.Text()
		// if !strings.HasPrefix(tt, "{") {
		tt = strings.TrimSpace(tt)
		tt = strings.ReplaceAll(tt, "'", "")
		tt = strings.ReplaceAll(tt, "`", "")
		tt = strings.ReplaceAll(tt, "\"", "")
		// }
		if strings.HasPrefix(tt, "#!(") {
			i := isolateFirstCommand([]byte(tt))
			logger.Printf("running subcommand '%s'\n", tt)
			tt = *EvalMacro([]byte(tt)[i[0]:i[1]], macros)
		}
		tt = strings.TrimPrefix(tt, "{")
		tt = strings.TrimSuffix(tt, "}")

		if tt == "" {
			continue
		}
		logger.Printf("Got token - '%s'\n", tt)
		args = append(args, tt)
	}
	var res string = fmt.Sprintf("`Invalid command: '%s'`\n", strings.Join(args, " "))
	for _, macro := range macros {
		if args[0] == macro.MacroName {
			res = macro.MacroFunc(&args)
			break
		}
	}
	return &res

}

func isolateFirstCommand(command []byte) []int {
	i := command_selector.FindIndex(command)
	if i == nil {
		return nil
	}
	var parenCollection []paren
	for _, j := range paren_selector.FindAllIndex(command, -1) {
		parenCollection = append(parenCollection, paren{index: j[0], isOpen: ((command)[j[0]] == '(')})
	}
	index := 1
	parenC := 1
	for parenC > 0 {
		if index >= len(parenCollection) {
			// line, col := getLineAndColNumbersFromIndex(&command, parenCollection[0].index)
			// logger.Printf("Unmatched parentheses starting at [%d:%d]\n", line, col)
			return nil
		}
		if parenCollection[index].isOpen {
			parenC++
		} else {
			parenC--
		}
		index++
	}
	return []int{i[0], parenCollection[index-1].index + 1}
}
func isolateCurlyBracketSet(command []byte) []int {
	i := start_curly.FindIndex(command)
	if i == nil {
		return nil
	}
	var parenCollection []paren
	for _, j := range curly_bracket_selector.FindAllIndex(command, -1) {
		parenCollection = append(parenCollection, paren{index: j[0], isOpen: ((command)[j[0]] == '{')})
	}
	index := 1
	parenC := 1
	for parenC > 0 {
		if index >= len(parenCollection) {
			// line, col := getLineAndColNumbersFromIndex(&command, parenCollection[0].index)
			// logger.Printf("Unmatched parentheses starting at [%d:%d]\n", line, col)
			return nil
		}
		if parenCollection[index].isOpen {
			parenC++
		} else {
			parenC--
		}
		index++
	}
	return []int{i[0], parenCollection[index-1].index + 1}
}
func isolateBracketSet(command []byte) []int {
	i := start_paren.FindIndex(command)
	if i == nil {
		return nil
	}
	var parenCollection []paren
	for _, j := range paren_selector.FindAllIndex(command, -1) {
		parenCollection = append(parenCollection, paren{index: j[0], isOpen: ((command)[j[0]] == '(')})
	}
	index := 1
	parenC := 1
	for parenC > 0 {
		if index >= len(parenCollection) {
			// line, col := getLineAndColNumbersFromIndex(&command, parenCollection[0].index)
			// logger.Printf("Unmatched parentheses starting at [%d:%d]\n", line, col)
			return nil
		}
		if parenCollection[index].isOpen {
			parenC++
		} else {
			parenC--
		}
		index++
	}
	return []int{i[0], parenCollection[index-1].index + 1}
}
func inlineReplaceSubstring(str *string, oldIndexes []int, newString string) {
	oldEnd := (*str)[oldIndexes[1]:]
	*str = (*str)[:oldIndexes[0]] + newString + oldEnd
}
func inlineReplaceBytes(data *[]byte, oldIndexes []int, newData *[]byte) {
	oldEnd := (*data)[(oldIndexes)[1]:]
	*data = append((*data)[:(oldIndexes)[0]], append((*newData), oldEnd...)...)
}

func (p *Parser) RegisterMacro(macro *Macro) {
	p.macros = append(p.macros, macro)
}
func (p *Parser) RegisterMacros(macros []*Macro) {
	p.macros = append(p.macros, macros...)
}

func getLineAndColNumbersFromIndex(content *[]byte, index int) (int, int) {
	lines := bytes.Split((*content)[:index+1], []byte("\n"))
	sum := 0
	for _, line := range lines[:len(lines)-1] {
		sum += len(line) + 1
		// logger.Printf("line: %s\n", line)
	}
	return len(lines), index - sum + 1
}
