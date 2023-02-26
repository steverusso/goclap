package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

const (
	typBool = iota
	typString
)

type buildVersionInfo struct {
	modVersion      string
	commitHash      string
	commitDate      string
	hasLocalChanges bool
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

type clapData struct {
	blurb    string
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
	data        clapData
	Opts        []optInfo
	Args        []argInfo
	Subcmds     []command
}

type optInfo struct {
	fieldType int
	FieldName string
	Long      string
	Short     string
	data      clapData
}

type argInfo struct {
	fieldType int
	FieldName string
	name      string
	data      clapData
}

func (c *command) DocName() string {
	if cfgName, ok := c.data.getConfig("cmd_name"); ok {
		return cfgName
	}
	return strings.ToLower(c.FieldName)
}

func run(rootCmdTypeName, srcDir string) error {
	fset := token.NewFileSet() // positions are relative to fset
	parsedDir, err := parser.ParseDir(fset, srcDir, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	b := builder{pkg: parsedDir["main"]}

	data := b.getCmdClapData(rootCmdTypeName)
	if data.blurb == "" {
		warn("no root command description provided\n")
	}
	rootStrct := b.findStruct(rootCmdTypeName)
	if rootStrct == nil {
		return fmt.Errorf("could not find a struct type named '%s'", rootCmdTypeName)
	}
	root := command{
		TypeName: rootCmdTypeName,
		// This is a bit of a hack due to the following: the "name" of the root command is
		// actually the name of the program, which is the first argument in `os.Args`.
		// That gets passed as a fmt arg within the generated code when printing a
		// command's usage. Therefore, we need a `%s` to show up wherever the root command
		// name will appear in a usage message.
		FieldName: "%[1]s",
		data:      data,
	}
	if err := b.addChildren(&root, rootStrct); err != nil {
		return err
	}

	outName := "./clap.go"
	f, err := os.OpenFile(outName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("opening '%s': %w", outName, err)
	}
	defer f.Close()

	g, err := newGenerator(f)
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
	return nil
}

func warn(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "\033[1;33mwarning:\033[0m "+format, a...)
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintln(os.Stderr, "error: no root command type provided")
		fmt.Fprintln(os.Stderr, "usage: goclap <type>")
		os.Exit(1)
	}
	typeName := os.Args[1]
	srcDir := "."
	if err := run(typeName, srcDir); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}
