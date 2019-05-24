package main

import (
	"flag"
	"github.com/emicklei/dot"
	"github.com/pkg/browser"
	"go/build"
	"log"
	"os"
	"os/exec"
	"strconv"
)

var (
	pkgDependences = make(map[string][]string)
)

func importAnalysis(pkg string) {

	// empty pkg, ignore it
	if len(pkg) == 0 {
		return
	}

	// pkg already analysed, ignore it
	if _, ok := pkgDependences[pkg]; ok {
		return
	}

	// fake package 'C', ignore it. https://golang.org/cmd/cgo/
	if pkg == "C" {
		return
	}

	importPkg, err := build.Import(pkg, "", 0)
	if err != nil {
		log.Fatal(err)
	}

	if len(importPkg.Imports) == 0 {
		return
	}

	pkgDependences[pkg] = importPkg.Imports
	for _, pkg := range importPkg.Imports {
		importAnalysis(pkg)
	}
}

func checkError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func formatDot() []byte {
	g := dot.NewGraph(dot.Directed)
	for pkg, dependencies := range pkgDependences {
		zero := g.Node(pkg)
		setNodeFontsize(&zero)
		for _, dependence := range dependencies {
			one := g.Node(dependence)
			setNodeFontsize(&one)
			g.Edge(zero, one)
		}
	}
	return []byte(g.String())
}

// 根据出现次数， 调节字体大小， 出现次数越多， 字体越大
func setNodeFontsize(zero *dot.Node) {
	if weight := zero.Value("fontsize"); weight != nil {
		if sw, ok := weight.(string); ok {
			iw, _ := strconv.Atoi(sw)
			zero.Attr("fontsize", strconv.Itoa(iw+3))
		}
	} else {
		zero.Attr("fontsize", "10")
	}
}

func main() {
	flag.Parse()
	pkgs := flag.Args()

	// prepare data
	for _, pkg := range pkgs {
		importAnalysis(pkg)
	}
	bformated := formatDot()

	// output data
	cmd := exec.Command("dot", "-Tsvg")
	in, err := cmd.StdinPipe()
	checkError(err)
	out, err := cmd.StdoutPipe()
	checkError(err)
	cmd.Stderr = os.Stderr

	checkError(cmd.Start())

	_, err = in.Write(bformated)
	checkError(err)

	checkError(in.Close())

	// show data
	ech := make(chan error)
	go func() {
		ech <- browser.OpenReader(out)
	}()

	checkError(cmd.Wait())
	checkError(<-ech)
}
