package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

//go:generate goclap -type goclap

// Pre-build tool to generate command line argument parsing code from Go comments.
type goclap struct {
	// Print version info and exit.
	//
	// clap:opt version,v
	version bool
	// Include the version info in the generated code.
	//
	// clap:opt include-version
	incVersion bool
	// The root command struct name.
	//
	// clap:opt type
	rootCmdType string
	// Directory of source files to parse (default ".").
	//
	// clap:opt srcdir
	srcDir string
	// Output file path (default "./clap.gen.go").
	//
	// clap:opt out,o
	outFilePath string
}

type basicType string

func (t basicType) IsBool() bool { return t == "bool" }

type buildVersionInfo struct {
	modVersion      string
	commitHash      string
	commitDate      string
	hasLocalChanges bool
}

func getBuildVersionInfo() buildVersionInfo {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		warn("unable to read build info")
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
	overview []string // paragraphs
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
	data      clapData
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

	rootCmd, pkgName, err := parse(c.srcDir, rootCmdTypeName)
	if err != nil {
		return err
	}

	code, err := generate(c.incVersion, pkgName, &rootCmd)
	if err != nil {
		return err
	}

	if c.outFilePath == "" {
		c.outFilePath = "./clap.gen.go"
	}
	f, err := os.OpenFile(c.outFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("opening '%s': %w", c.outFilePath, err)
	}
	defer f.Close()

	if _, err = f.Write(code); err != nil {
		return fmt.Errorf("writing code to output file: %w", err)
	}

	return nil
}

func warn(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "\033[1;33mwarning:\033[0m "+format+"\n", a...)
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
