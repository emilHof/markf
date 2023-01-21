package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/OutboundSpade/markf/logger"
	"github.com/OutboundSpade/markf/macro"

	pdfgen "github.com/OutboundSpade/markf/pdf_gen"
)

func main() {
	var outfile string
	var printOutput bool
	var allowUnsafe bool
	var enableLogging bool
	flag.StringVar(&outfile, "o", "", "output file")
	flag.BoolVar(&printOutput, "p", false, "print output to stdout")
	flag.BoolVar(&allowUnsafe, "allow-unsafe", false, "allow unsafe macros")
	flag.BoolVar(&enableLogging, "d", false, "enable debug logging")
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] <input file>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	logger.ENABLE_LOGGING = enableLogging
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	file, err := os.Open(flag.Arg(0))
	must(err)
	defer file.Close()

	//get just file name
	fName := strings.Split(flag.Arg(0), "/")[len(strings.Split(flag.Arg(0), "/"))-1]
	//get just file name without fExt
	fName = strings.Split(fName, ".")[0]
	//set output file name
	if outfile == "" {
		outfile = fmt.Sprintf("%s.pdf", fName)
	}

	doc, err := io.ReadAll(file)
	must(err)

	p := macro.Parser{}
	p.RegisterMacros(macro.DefaultMacros)
	if allowUnsafe {
		p.RegisterMacros(macro.UnsafeMacros)
	}
	loadMacros(&p)
	p.Process(&doc)

	if printOutput {
		fmt.Println(string(doc))
	} else if strings.HasSuffix(outfile, ".pdf") {
		pdf, err := pdfgen.RenderPDF(&doc)
		must(err)
		(*pdf).WritePdf(outfile)
	} else if strings.HasSuffix(outfile, ".md") {
		file, err := os.Create(outfile)
		must(err)
		defer file.Close()
		_, err = file.Write(doc)
		must(err)
	} else {
		panic("Unknown file type")
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
