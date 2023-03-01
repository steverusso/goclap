package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

//go:generate goclap goclap

// pre-build tool to generate command line argument parsing code from Go comments.
type goclap struct {
	// print version info and exit
	//
	// clap:opt version,v
	version bool
	// the root command struct name
	//
	// clap:opt type
	rootCmdType string
	// directory of source files to parse (default ".")
	//
	// clap:opt srcdir
	srcDir string
	// output file path (default "./clap.go")
	//
	// clap:opt out
	outFilePath string
}

type basicType int

const (
	typBool basicType = iota
	typString
)

func (t basicType) IsBool() bool   { return t == typBool }
func (t basicType) IsString() bool { return t == typString }

type buildVersionInfo struct {
	modVersion      string
	commitHash      string
	commitDate      string
	hasLocalChanges bool
}

func getBuildVersionInfo() buildVersionInfo {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		warn("unable to read build info\n")
		return buildVersionInfo{}
	}
	v := buildVersionInfo{
		modVersion: bi.Main.Version,
	}
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			v.commitHash = s.Value
		case "vcs.time":
			t, err := time.Parse(time.RFC3339, s.Value)
			if err != nil {
				warn("unable to parse vcs.time '%s': %v", s.Value, err)
			} else {
				v.commitDate = t.Format("20060102")
			}
		case "vcs.modified":
			v.hasLocalChanges = (s.Value == "true")
		}
	}
	return v
}

func (v buildVersionInfo) String() string {
	hash := v.commitHash
	if len(hash) > 12 {
		hash = v.commitHash[:12]
	}
	s := v.modVersion
	if v.commitDate != "" {
		s += "-" + v.commitDate
	}
	if hash != "" {
		s += "-" + hash
	}
	if v.hasLocalChanges {
		s += "-(with unstaged changes)"
	}
	return s
}

type clapData struct {
	Blurb    string
	longDesc string
	configs  []clapConfig
}

type clapConfig struct {
	key string
	val string
}

func (d *clapData) getConfig(k string) (string, bool) {
	for i := range d.configs {
		if k == d.configs[i].key {
			return d.configs[i].val, true
		}
	}
	return "", false
}

type command struct {
	parentNames []string
	FieldName   string
	TypeName    string
	Data        clapData
	Opts        []option
	Args        []argument
	Subcmds     []command
}

type option struct {
	FieldType basicType
	FieldName string
	Long      string
	Short     string
	data      clapData
}

type argument struct {
	FieldType basicType
	FieldName string
	name      string
	Data      clapData
}

func (c *command) UsgName() string {
	if cfgName, ok := c.Data.getConfig("cmd_name"); ok {
		return cfgName
	}
	return strings.ToLower(c.FieldName)
}

func gen(c *goclap) error {
	rootCmdTypeName := c.rootCmdType
	if rootCmdTypeName == "" {
		claperr("no root command type provided\n")
		c.printUsage(os.Stderr)
		os.Exit(1)
	}

	srcDir := "."
	if c.srcDir != "" {
		srcDir = c.srcDir
	}
	fset := token.NewFileSet() // positions are relative to fset
	parsedDir, err := parser.ParseDir(fset, srcDir, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	b := builder{pkg: parsedDir["main"]}

	rootStrct := b.findStruct(rootCmdTypeName)
	if rootStrct == nil {
		return fmt.Errorf("could not find a struct type named '%s'", rootCmdTypeName)
	}
	data := b.getCmdClapData(rootCmdTypeName)
	if data.Blurb == "" {
		warn("no root command description provided\n")
	}
	root := command{
		TypeName: rootCmdTypeName,
		// This is a bit of a hack due to the following: the "name" of the root command is
		// actually the name of the program, which is the first argument in `os.Args`.
		// That gets passed as a fmt arg within the generated code when printing a
		// command's usage. Therefore, we need a `%s` to show up wherever the root command
		// name will appear in a usage message.
		FieldName: "%[1]s",
		Data:      data,
	}
	if err := b.addChildren(&root, rootStrct); err != nil {
		return err
	}

	g, err := newGenerator()
	if err != nil {
		return fmt.Errorf("initializing generator: %w", err)
	}
	err = g.writeHeader(root.HasSubcmds())
	if err != nil {
		return err
	}
	if err := g.generate(&root); err != nil {
		return fmt.Errorf("generating: %w", err)
	}

	outName := "./clap.go"
	if c.outFilePath != "" {
		outName = c.outFilePath
	}
	f, err := os.OpenFile(outName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("opening '%s': %w", outName, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, &g.buf); err != nil {
		return fmt.Errorf("copying buffer to output file: %w", err)
	}
	return nil
}

func warn(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "\033[1;33mwarning:\033[0m "+format, a...)
}

func main() {
	c := goclap{}
	c.parse(os.Args)

	if c.version {
		fmt.Println(getBuildVersionInfo())
		return
	}

	if err := gen(&c); err != nil {
		claperr("%v\n", err)
		os.Exit(1)
	}
}
