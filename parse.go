package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"
)

// backtickRepl describes how groups of backticks are replaced within usage message
// strings using regular expressions. The usage messages are raw strings in the generated
// code, so backticks would be syntactically broken delimiters of different raw strings.
// Therefore, groups of backticks are placed into their own double-quoted strings and
// concatenated to the rest of the usage message string.
const backtickRepl = "`+\"$0\"+`"

// One or more backticks.
var backtickRE = regexp.MustCompile("`+")

// helpOption is the default help option that is automatically added to any command's
// options.
var helpOption = option{
	Short:     "h",
	Long:      "help",
	FieldType: "bool",
	data:      clapData{Blurb: "show this help message"},
}

func parse(srcDir, rootCmdTypeName string) (command, string, error) {
	if srcDir == "" {
		srcDir = "."
	}

	fset := token.NewFileSet() // positions are relative to fset
	parsedPkgs, err := parser.ParseDir(fset, srcDir, nil, parser.ParseComments)
	if err != nil {
		return command{}, "", err
	}

	var targetPkg *ast.Package
	var rootStrct *ast.StructType
	for _, pkg := range parsedPkgs {
		if s := findStruct(pkg, rootCmdTypeName); s != nil {
			targetPkg = pkg
			rootStrct = s
			break
		}
	}
	if targetPkg == nil {
		return command{}, "", fmt.Errorf("could not find a struct type named '%s'", rootCmdTypeName)
	}

	data := getCmdClapData(targetPkg, rootCmdTypeName)
	if data.Blurb == "" {
		warn("no root command description provided")
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

	if err = addChildren(targetPkg, &root, rootStrct); err != nil {
		return command{}, "", err
	}
	return root, targetPkg.Name, nil
}

func addChildren(pkg *ast.Package, c *command, strct *ast.StructType) error {
	// Read in the struct fields.
	for _, field := range strct.Fields.List {
		if len(field.Names) > 1 {
			warn("skipping multi named field %s", field.Names)
			continue
		}
		fieldName := field.Names[0].Name
		typeAndField := fmt.Sprintf("'%s.%s'", c.TypeName, fieldName)
		if _, ok := field.Type.(*ast.StructType); ok {
			warn("skipping %s (commands must be struct pointers)", typeAndField)
			continue
		}
		if star, ok := field.Type.(*ast.StarExpr); ok {
			idnt, ok := star.X.(*ast.Ident)
			if !ok {
				warn("skipping %s: non-struct pointers are unsupported", typeAndField)
				continue
			}
			// The field, which is of type `*IDENT,` will be a command if `IDENT`
			// identifies a struct defined within this package.
			subStrct := findStruct(pkg, idnt.Name)
			if subStrct == nil {
				warn("skipping %s: if type '%s' is defined, it's not a struct", typeAndField, idnt.Name)
				continue
			}
			// The field is firmly considered a subcommand at this point.
			subcmd := command{
				parentNames: append(c.parentNames, c.UsgName()),
				TypeName:    idnt.Name,
				FieldName:   fieldName,
				Data:        getCmdClapData(pkg, idnt.Name),
			}
			// Recursively build this subcommand from it's own struct type definition.
			err := addChildren(pkg, &subcmd, subStrct)
			if err != nil {
				return err
			}
			c.Subcmds = append(c.Subcmds, subcmd)
			continue
		}
		// From now on, it's either an option or an argument which can only be basic types
		// (and those start out as identifiers).
		idnt, ok := field.Type.(*ast.Ident)
		if !ok {
			warn("skipping %s (looking for ident, unsure how to handle %T)", typeAndField, field.Type)
			continue
		}
		fieldType := basicTypeFromName(idnt.Name)
		if fieldType == "" {
			warn("skipping %s: unsupported option or argument type '%s'", typeAndField, idnt.Name)
			continue
		}
		fieldDocs := parseComments(field.Doc)
		cfgTypes := scanConfigTypes(fieldDocs.configs)
		if cfgTypes.opts {
			if cfgTypes.args {
				return fmt.Errorf("%s has both option and argument configurations", typeAndField)
			}
			// The field is firmly considered an option at this point.
			err := c.addOption(fieldDocs, fieldName, fieldType)
			if err != nil {
				return fmt.Errorf("parsing %s field as option: %w", typeAndField, err)
			}
			continue
		}
		// The field is assumed to be an argument at this point.
		if fieldType.IsBool() {
			return fmt.Errorf("%s: arguments cannot be type bool", typeAndField)
		}
		c.Args = append(c.Args, argument{
			data:      fieldDocs,
			FieldType: fieldType,
			FieldName: fieldName,
			name:      strings.ToLower(fieldName),
		})
	}
	c.Opts = append(c.Opts, helpOption)
	return nil
}

type cfgTypes struct {
	opts bool
	args bool
}

func scanConfigTypes(cfgs []clapConfig) cfgTypes {
	var ct cfgTypes
	for i := range cfgs {
		k := cfgs[i].key
		if strings.HasPrefix(k, "opt") {
			ct.opts = true
		}
		if strings.HasPrefix(k, "arg") {
			ct.args = true
		}
	}
	return ct
}

func basicTypeFromName(name string) basicType {
	switch name {
	case "bool", "string", "byte", "rune", "float32", "float64",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return basicType(name)
	}
	return ""
}

func findStruct(pkg *ast.Package, name string) *ast.StructType {
	var strct *ast.StructType
	ast.Inspect(pkg, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.GenDecl:
			for i := range n.Specs {
				if s, ok := n.Specs[i].(*ast.TypeSpec); ok && s.Name.Name == name {
					strct = s.Type.(*ast.StructType)
					return false
				}
			}
		case *ast.TypeSpec:
			if n.Name.Name == name && n.Doc != nil {
				strct = n.Type.(*ast.StructType)
				return false
			}
		}
		return true
	})
	return strct
}

func getCmdClapData(pkg *ast.Package, typ string) clapData {
	var commentGrp *ast.CommentGroup
	ast.Inspect(pkg, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.GenDecl:
			if n.Doc != nil {
				for i := range n.Specs {
					if s, ok := n.Specs[i].(*ast.TypeSpec); ok && s.Name.Name == typ {
						commentGrp = n.Doc
						return false
					}
				}
			}
		case *ast.TypeSpec:
			if n.Name.Name == typ && n.Doc != nil {
				commentGrp = n.Doc
				return false
			}
		}
		return true
	})
	return parseComments(commentGrp)
}

func parseOptNames(str string) (string, string, error) {
	names := strings.Split(str, ",")
	for i := len(names) - 1; i >= 0; i-- {
		if names[i] == "" {
			names = append(names[:i], names[i+1:]...)
		}
	}
	long := ""
	short := ""
	switch len(names) {
	case 0:
		return "", "", errors.New("'clap:opt' found but no names provided")
	case 1:
		v := names[0]
		if len(v) == 1 {
			short = v
		} else {
			long = v
		}
	case 2:
		a, b := names[0], names[1]
		if len(a) == 1 {
			long, short = b, a
		} else if len(b) == 1 {
			long, short = a, b
		} else {
			return "", "", fmt.Errorf("two opt names found ('%s', '%s'), one must be the short version (only one character)", a, b)
		}
	default:
		return "", "", fmt.Errorf("illegal `clap:opt` value '%s': too many comma separated values", str)
	}
	if long == "help" || short == "h" {
		return "", "", errors.New("'help' and 'h' are reserved option names")
	}
	return long, short, nil
}

func (c *command) addOption(data clapData, fieldName string, typ basicType) error {
	names, ok := data.getConfig("opt")
	if !ok {
		return errors.New("adding option without a 'clap:opt' directive")
	}
	long, short, err := parseOptNames(names)
	if err != nil {
		return fmt.Errorf("parsing option names: %w", err)
	}
	c.Opts = append(c.Opts, option{
		FieldType: typ,
		FieldName: fieldName,
		Long:      long,
		Short:     short,
		data:      data,
	})
	return nil
}

func parseComments(cg *ast.CommentGroup) clapData {
	if cg == nil {
		return clapData{}
	}

	cd := clapData{}

	lines := strings.Split(cg.Text(), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.HasPrefix(lines[i], "clap:") {
			rest := lines[i]
			rest = rest[len("clap:"):]
			var j int
			for j = 0; j < len(rest); j++ {
				if rest[j] == ' ' {
					break
				}
			}
			cfg := clapConfig{
				key: rest[:j],
			}
			if j < len(rest) {
				cfg.val = rest[j+1:]
			}
			cd.configs = append([]clapConfig{cfg}, cd.configs...)
			lines = append(lines[:i], lines[i+1:]...)
		}
	}

	// Grab all lines up to the first blank one as the "blurb."
	for i := range lines {
		if lines[i] == "" {
			cd.Blurb = strings.TrimSpace(strings.Join(lines[:i], " "))
			lines = lines[i+1:]
			break
		}
	}

	// Drop trailing '.' punctuation.
	if n := len(cd.Blurb); n > 0 && cd.Blurb[n-1] == '.' {
		cd.Blurb = cd.Blurb[:n-1]
	}

	// The remaining groups of non-empty lines (if any) are considered the paragraphs of
	// the item's "overview" (only ever used for commands, not for options or arguments).
	paras := make([]string, 0, 2)
	var p strings.Builder
	for i := range lines {
		if lines[i] != "" {
			p.WriteString(lines[i])
			p.WriteByte('\n')
		} else {
			if i > 0 && lines[i-1] != "" {
				paras = append(paras, p.String())
				p.Reset()
			}
		}
	}
	if p.Len() > 0 {
		paras = append(paras, p.String())
	}
	cd.overview = paras

	// Put groups of backticks into their own strings.
	cd.Blurb = backtickRE.ReplaceAllString(cd.Blurb, backtickRepl)
	for i := range cd.overview {
		cd.overview[i] = backtickRE.ReplaceAllString(cd.overview[i], backtickRepl)
	}

	return cd
}
