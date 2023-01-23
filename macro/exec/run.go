package exec

import (
	"bufio"
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/fogleman/gg"
)

var FG_COLOR = color.Black
var BG_COLOR = color.White

// var asni_capture = regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))")
var wierd = regexp.MustCompile("[^a-zA-Z0-9\n`~!@#$%^&*()\\[\\]{}\\\\\\<\\>|_+-= \t]+")
var dup = regexp.MustCompile("[\\s\\S]*\a")

// var asni_capture = regexp.MustCompile("\x1B(?:[@-Z\\-_]|[[0-?]*[ -/]*[@-~])")
// var asni_capture = regexp.MustCompile("[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]")
const ENDC string = "endc"

func ScreenShotCommands(commands []string) (*[]byte, error) {
	shell_cmd := exec.Command("/bin/bash", "-i")
	shell_cmd.Env = os.Environ()
	shell_cmd.Env = append(shell_cmd.Env, "TERM=xterm")
	var full []byte

	_, stderr, stdin, err := getIOPiped(shell_cmd)
	if err != nil {
		return nil, err
	}

	var shell_txt []byte
	//run
	if err = shell_cmd.Start(); err != nil {
		return nil, err
	}

	var tmp []byte
	if tmp, err = readUntil(stderr, "$ "); err != nil {
		return nil, err
	}
	tmp = dup.ReplaceAll(tmp, []byte(""))
	shell_txt = append(shell_txt, tmp...)
	shell_txt = wierd.ReplaceAll(shell_txt, []byte(""))
	shell_txt = []byte(strings.TrimSpace(string(shell_txt)))
	shell_txt = append(shell_txt, []byte(" ")...)

	if _, err = stdin.Write([]byte("exit\n")); err != nil {
		return nil, err
	}
	shell_cmd.Wait()

	full = append(full, shell_txt...)
	for _, command := range commands {
		full = append(full, (command + "\n")...)
		full = append(full, *RunCommand(command)...)
		// }
		full = append(full, shell_txt...)
	}
	out, err := text2Png(string(full))
	if err != nil {
		return nil, err
	}
	return out, nil

}

func RunCommand(command string) *[]byte {
	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.Env = os.Environ()
	out, _ := cmd.CombinedOutput()
	// out = out[:len(out)-1]
	return &out
}

var WIDTH = 600
var LINE_SPACING = 1.5

func text2Png(text string) (*[]byte, error) {
	// var h = 400.0
	ctx1 := gg.NewContext(WIDTH, 600)
	lc := drawStringWrapped(ctx1, text, LINE_SPACING*2, LINE_SPACING*2, float64(WIDTH), LINE_SPACING)
	ctx := gg.NewContext(WIDTH, int(float64(lc+1)*(LINE_SPACING*float64(ctx1.FontHeight()))))
	ctx.SetColor(BG_COLOR)
	ctx.Clear()
	ctx.SetColor(FG_COLOR)
	// ctx.DrawStringWrapped(text, LINE_SPACING*2, 0, 0, 0, float64(WIDTH), LINE_SPACING, 0)

	drawStringWrapped(ctx, text, 0, 0, float64(WIDTH), LINE_SPACING)

	//new writer
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	err := png.Encode(w, ctx.Image())
	if err != nil {
		return nil, err
	}
	w.Flush()
	out := buf.Bytes()
	// f, _ := os.OpenFile("out.png", os.O_CREATE|os.O_WRONLY, 0644)
	// f.Write(out)
	return &out, nil
}

// returns the number of lines
func drawStringWrapped(ctx *gg.Context, text string, x, y, width, lineSpacing float64) int {
	lines := strings.Split(text, "\n")
	lineCount := 0
	for _, rawLine := range lines {
		lineWidth, _ := ctx.MeasureString(rawLine)
		splitLine := []string{}
		if lineWidth > width {
			words := strings.Split(rawLine, " ")
			line := ""
			for wi, word := range words {
				wordWidth, _ := ctx.MeasureString(word)
				if wordWidth > width {
					// word is too long to fit on a line
					// split it up
					for _, char := range word {
						if wordWidth, _ = ctx.MeasureString(line + string(char)); wordWidth > width {
							splitLine = append(splitLine, line)
							line = ""
						}
						line += string(char)
						// splitLine = append(splitLine, word[i:i+int(width)])
					}
				} else if wordWidth, _ = ctx.MeasureString(line + " " + word); wordWidth > width {
					// word is too long to fit on this line
					// add it to the next line
					splitLine = append(splitLine, line)
					line = word + " "
				} else {
					// word fits on this line
					line += word + " "
				}
				if wi == len(words)-1 {
					splitLine = append(splitLine, line)
				}
			}
		} else {
			splitLine = append(splitLine, rawLine)
		}
		for _, line := range splitLine {
			lineCount += 1
			fmt.Printf("'%s'\n", line)
			ctx.DrawString(line, x, y+float64(lineCount)*LINE_SPACING*ctx.FontHeight())
			// ctx.DrawString(line, LINE_SPACING*2, float64(lineCount)*LINE_SPACING*ctx.FontHeight())
			// ctx.DrawStringAnchored(line, LINE_SPACING*2, float64(lineCount)*LINE_SPACING, 0, float64(lineCount)*LINE_SPACING+1)
		}
	}
	return lineCount
}

func readUntil(r io.Reader, delim string) ([]byte, error) {
	var buf []byte
	for {
		bScan := make([]byte, len(delim))
		_, err := r.Read(bScan)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		buf = append(buf, bScan...)
		if string(bScan) == delim {
			break
		}
	}
	return buf, nil
}
func getIOPiped(command *exec.Cmd) (stdout io.ReadCloser, stderr io.ReadCloser, stdin io.WriteCloser, err error) {
	stdout, err = command.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	stderr, err = command.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	stdin, err = command.StdinPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	return stdout, stderr, stdin, nil
}
