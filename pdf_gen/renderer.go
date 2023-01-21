package pdfgen

import (
	"embed"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/OutboundSpade/markf/logger"
	"github.com/fogleman/gg"
	"github.com/russross/blackfriday/v2"
	"github.com/signintech/gopdf"
)

//go:embed fonts/*
var fonts embed.FS

const FONT_DIR = "fonts"
const FONT_EXTENSION = ".ttf"

const MARGIN = 50
const PAGE_WIDTH = 595.28
const PAGE_HEIGHT = 841.89

func RenderPDF(doc *[]byte) (*gopdf.GoPdf, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: PAGE_WIDTH, H: PAGE_HEIGHT}})
	pdf.AddPage()
	pdf.SetMargins(MARGIN, MARGIN, MARGIN, MARGIN)
	pdf.Br(0)
	var err error

	fontFiles, err := fonts.ReadDir(FONT_DIR)
	if err != nil {
		return nil, err
	}
	for _, fontFileEntry := range fontFiles {
		extension := fontFileEntry.Name()[len(fontFileEntry.Name())-4:]
		fname := fontFileEntry.Name()[:len(fontFileEntry.Name())-4]
		if extension == FONT_EXTENSION {
			fontFile, err := fonts.Open(FONT_DIR + "/" + fontFileEntry.Name())
			if err != nil {
				return nil, err
			}
			style := gopdf.Regular
			if strings.Contains(fname, "-Bold") {
				style = gopdf.Bold
				fname = strings.ReplaceAll(fname, "-Bold", "")
			} else if strings.Contains(fname, "-Italic") {
				style = gopdf.Italic
				fname = strings.ReplaceAll(fname, "-Italic", "")
			}
			err = pdf.AddTTFFontByReaderWithOption(fname, fontFile, gopdf.TtfOption{Style: style})
			if err != nil {
				return nil, err
			}
		}
	}
	// err = pdf.SetFont("wts11", "", FONT_SIZE)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	md := blackfriday.New()
	node := md.Parse(*doc)
	parseNode(node, &pdf, 0)
	return &pdf, nil
}

const FONT_SIZE = 14
const LINE_HEIGHT = 1.5

var TEXT_COLOR = []uint8{0, 0, 0} // in rgb
const BREAK_HEIGHT = LINE_HEIGHT * FONT_SIZE

var DEFAULT_CELL_RECT = &gopdf.Rect{W: PAGE_WIDTH - MARGIN*2, H: BREAK_HEIGHT * 1.1}
var CELL_RECT *gopdf.Rect = nil

const TOP_HEADER_FONT_SIZE = 28
const MAIN_FONT = "Roboto"
const CODE_FONT = "Courier"

var defaultCellOptions = gopdf.CellOption{Align: gopdf.Center | gopdf.Middle}
var cellOptions = defaultCellOptions

func parseNode(node *blackfriday.Node, pdf *gopdf.GoPdf, level int) {
	if pdf.GetY() > PAGE_HEIGHT-MARGIN*2 {
		pdf.AddPage()
	}
	switch node.Type {
	case blackfriday.Heading:
		//grows by 2^(-x)*10
		const base = 1.5
		fSize := FONT_SIZE + ((math.Pow(base, float64(-(node.HeadingData.Level))) * (TOP_HEADER_FONT_SIZE - FONT_SIZE)) * base)
		// fmt.Println(fSize)
		pdf.SetFont(MAIN_FONT, "B", fSize)
		// pdf.Br(BREAK_HEIGHT/FONT_SIZE*fSize + 10)
		pdf.Br(BREAK_HEIGHT)
	case blackfriday.Paragraph:
		pdf.SetFont(MAIN_FONT, "", FONT_SIZE)
		if (node.Parent.Type == blackfriday.Item) || (node.Parent.Type == blackfriday.List) {
			// pdf.CellWithOption(CELL_RECT, strings.Repeat(" ", level*2))
		} else {
			pdf.Br(BREAK_HEIGHT)
		}
	case blackfriday.List:
		// pdf.Br(BREAK_HEIGHT)
		pdf.SetFont(MAIN_FONT, "", FONT_SIZE)
	case blackfriday.Item:
		// data := node
		var prefix string
		for i := 0; i < level; i++ {
			prefix += "  "
		}
		pdf.Br(BREAK_HEIGHT)
		drawTextWrapped(pdf, prefix+"• ", node.Parent)
		// pdf.CellWithOption(nil, prefix+"• ", cellOptions)
	case blackfriday.CodeBlock:
		pdf.SetFont(CODE_FONT, "", FONT_SIZE)
		drawTextWrapped(pdf, string(node.Literal), node.Parent)
		pdf.Br(BREAK_HEIGHT)
		pdf.SetFont(MAIN_FONT, "", FONT_SIZE)
	case blackfriday.BlockQuote:
		pdf.SetFont(MAIN_FONT, "I", FONT_SIZE)
		pdf.Br(BREAK_HEIGHT)
		drawTextWrapped(pdf, string(node.Literal), node.Parent)
		pdf.SetFont(MAIN_FONT, "", FONT_SIZE)
	case blackfriday.Code:
		pdf.SetFont(CODE_FONT, "", FONT_SIZE)
		drawTextWrapped(pdf, string(node.Literal), node.Parent)
		// pdf.Br(BREAK_HEIGHT)
		// code := string(node.Literal)
		// for i, ln := range strings.Split(code, "\n") {
		// 	if i > 0 {
		// 		pdf.Br(BREAK_HEIGHT)
		// 	}
		// 	pdf.CellWithOption(CELL_RECT, ln)
		// }
		pdf.SetFont(MAIN_FONT, "", FONT_SIZE)
	case blackfriday.Link:
		pdf.SetFont(MAIN_FONT, "U", FONT_SIZE)
		// pdf.Br(BREAK_HEIGHT)
		txtWid, _ := pdf.MeasureTextWidth(string(node.FirstChild.Literal))
		pdf.AddExternalLink(
			string(node.LinkData.Destination),
			pdf.GetX(),
			pdf.GetY(),
			txtWid,
			FONT_SIZE)
		pdf.SetFont(MAIN_FONT, "", FONT_SIZE)
	case blackfriday.Emph:
		pdf.SetFont(MAIN_FONT, "I", FONT_SIZE)
	case blackfriday.Strong:
		pdf.SetFont(MAIN_FONT, "B", FONT_SIZE)
	case blackfriday.Hardbreak:
		pdf.Br(BREAK_HEIGHT)
	case blackfriday.Softbreak:
		pdf.Br(BREAK_HEIGHT)
	case blackfriday.HorizontalRule:
		pdf.Br(BREAK_HEIGHT)
		pdf.Line(MARGIN, pdf.GetY()+BREAK_HEIGHT, PAGE_WIDTH-MARGIN, pdf.GetY()+BREAK_HEIGHT)
		pdf.Br(BREAK_HEIGHT)
	case blackfriday.Document:
		// do nothing
	case blackfriday.Image:
		var img image.Image
		if strings.HasPrefix(string(node.LinkData.Destination), "http") {
			img = getImageFromURL(string(node.LinkData.Destination))
		} else if strings.HasPrefix(string(node.LinkData.Destination), "data:image/png;base64,") {
			n, err := png.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(strings.TrimPrefix(string(node.LinkData.Destination), "data:image/png;base64,"))))
			if err != nil {
				img = getPlaceHolderImage(err.Error())
			} else {
				img = n
			}
		} else {
			f, err := os.Open(string(node.LinkData.Destination))
			if err != nil {
				img = getPlaceHolderImage(err.Error())
			} else {
				n, err := png.Decode(f)
				if err != nil {
					img = getPlaceHolderImage(err.Error())
				} else {
					img = n
				}
			}
		}
		var rect *gopdf.Rect = nil
		if CELL_RECT != nil {
			rect = &gopdf.Rect{
				W: float64(img.Bounds().Max.X),
				H: float64(img.Bounds().Max.Y),
			}
			if cellOptions.Align&gopdf.Center > 0 {
				pdf.SetX((PAGE_WIDTH / 2) - (float64(img.Bounds().Max.X) / 2))
			}
		}
		pdf.ImageFrom(img, pdf.GetX(), pdf.GetY(), rect)
		if rect != nil {
			pdf.Br((*rect).H)
		} else {
			pdf.Br(float64(img.Bounds().Max.Y) / 2)
		}
		// fmt.Println(node.LinkData.Destination)
	case blackfriday.Text:
		drawTextWrapped(pdf, string(node.Literal), node.Parent)
	case blackfriday.HTMLSpan:
		parseHTMLSpan(pdf, string(node.Literal))
	default:
		fmt.Printf("Unknown node type: %s\n", node.Type.String())
	}
	prefix := ""
	for i := 0; i < level; i++ {
		prefix += "  "
	}
	logger.Printf("%s%s: '%s'\n", prefix, node.Type.String(), strings.ReplaceAll(string(node.Literal), "\n", "\\n"))
	for child := node.FirstChild; child != nil; child = child.Next {
		parseNode(child, pdf, level+1)
	}
}

var colors = map[string][]int{
	"red":    {255, 0, 0},
	"orange": {255, 165, 0},
	"yellow": {255, 255, 0},
	"green":  {0, 255, 0},
	"blue":   {0, 0, 255},
	"purple": {128, 0, 128},
	"white":  {255, 255, 255},
	"black":  {0, 0, 0},
}

func parseHTMLSpan(pdf *gopdf.GoPdf, text string) {
	cmd := strings.Split(strings.Trim(text, " <>"), " ")
	if strings.HasPrefix(cmd[0], "/") {
		cmd[0] = cmd[0][1:]
		switch cmd[0] {
		case "color":
			pdf.SetTextColor(TEXT_COLOR[0], TEXT_COLOR[1], TEXT_COLOR[2])
		case "center":
			cellOptions = defaultCellOptions
			CELL_RECT = nil
		}
		return
	}
	switch cmd[0] {
	case "color":
		c := colors[cmd[1]]
		if c != nil {
			pdf.SetTextColor(uint8(c[0]), uint8(c[1]), uint8(c[2]))
			return
		}
		color := strings.Split(cmd[1], ",")
		r, err := strconv.Atoi(color[0])
		mustBeNumber(err, color[0], text)
		g, err := strconv.Atoi(color[1])
		mustBeNumber(err, color[1], text)
		b, err := strconv.Atoi(color[2])
		mustBeNumber(err, color[2], text)
		mustBeUint8(r, text)
		mustBeUint8(g, text)
		mustBeUint8(b, text)
		pdf.SetTextColor(uint8(r), uint8(g), uint8(b))
	case "center":
		cellOptions = gopdf.CellOption{
			Align: gopdf.Center | gopdf.Middle,
		}
		CELL_RECT = DEFAULT_CELL_RECT
	case "pagebreak":
		pdf.AddPage()
		// fmt.Println("Page break")
	}
}
func mustBeNumber(e error, num string, text string) {
	if e != nil {
		panic(fmt.Sprintf("Error in '%s': '%s' is not a valid number\n%s", text, num, e))
	}
}
func mustBeUint8(i int, text string) {
	if i < 0 || i > 255 {
		panic(fmt.Sprintf("Error in '%s': '%d' is not a valid uint8 value", text, i))
	}
}

func drawTextWrapped(pdf *gopdf.GoPdf, text string, parent *blackfriday.Node) {
	txt, _ := pdf.SplitTextWithWordWrap(text, PAGE_WIDTH-MARGIN*2)
	if text == "" || text == "\n" {
		pdf.Br(BREAK_HEIGHT)
		return
	}
	for index, line := range txt {
		if index > 0 {
			pdf.Br(BREAK_HEIGHT)
		}
		if parent.Type == blackfriday.Link {
			pdf.SetTextColor(0, 0, 255)
			pdf.CellWithOption(nil, line, cellOptions)
			pdf.SetTextColor(TEXT_COLOR[0], TEXT_COLOR[1], TEXT_COLOR[2])
		} else {
			pdf.CellWithOption(CELL_RECT, line, cellOptions)
		}
		// pdf.Br(BREAK_HEIGHT / 2)
	}
	if parent.Type == blackfriday.Heading {
		pdf.Br(BREAK_HEIGHT * 0.5)
	}
	pdf.SetFont(MAIN_FONT, "", FONT_SIZE)
	// txt, _ := pdf.SplitTextWithWordWrap(text, PAGE_WIDTH-MARGIN*2)
	// out := strings.Join(txt, "\n")
	// if out == "\n" {
	// 	pdf.Br(BREAK_HEIGHT)
	// 	return
	// }
	// for i, ln := range strings.Split(out, "\n") {
	// 	if i > 0 {
	// 		pdf.Br(BREAK_HEIGHT)
	// 	}
	// 	pdf.CellWithOption(CELL_RECT, ln)
	// }

}

func getImageFromURL(url string) image.Image {
	httpResp, err := http.Get(url)
	if err != nil {
		return getPlaceHolderImage(err.Error())
	}
	if httpResp.StatusCode != 200 {
		return getPlaceHolderImage(httpResp.Status)
	}
	defer httpResp.Body.Close()
	img, err := png.Decode(httpResp.Body)
	if err != nil {

		return getPlaceHolderImage(err.Error())
	}
	return img
}

func getPlaceHolderImage(text string) image.Image {
	ctx := gg.NewContext(800, 150)
	ctx.SetHexColor("#C00000")
	ctx.Clear()
	ctx.SetHexColor("#FFFFFF")
	ctx.DrawStringWrapped(text, 0, 0, 0, 0, 800, 2, gg.AlignCenter)
	return ctx.Image()
}
