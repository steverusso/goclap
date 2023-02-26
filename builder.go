package main

import (
	"errors"
	"fmt"
	"go/ast"
	"strings"
)

var helpOption = optInfo{
	Short: "h",
	Long:  "help",
	Data:  clapData{Blurb: "show this help message"},
}

type builder struct {
	pkg *ast.Package
}

func (b *builder) addChildren(c *command, strct *ast.StructType) error {
	// Read in the struct fields.
	for _, field := range strct.Fields.List {
		if len(field.Names) > 1 {
			warn("skipping multi named field %s\n", field.Names)
			continue
		}
		fieldName := field.Names[0].Name
		typeAndField := fmt.Sprintf("'%s.%s'", c.TypeName, fieldName)
		if _, ok := field.Type.(*ast.StructType); ok {
			warn("skipping %s (commands must be struct pointers)\n", typeAndField)
			continue
		}
		if star, ok := field.Type.(*ast.StarExpr); ok {
			idnt, ok := star.X.(*ast.Ident)
			if !ok {
				warn("skipping %s: non-struct pointers are unsupported\n", typeAndField)
				continue
			}
			// The field, which is of type `*IDENT,` will be a command if `IDENT`
			// identifies a struct defined within this package.
			subStrct := b.findStruct(idnt.Name)
			if subStrct == nil {
				warn("skipping %s: if type '%s' is defined, it's not a struct", typeAndField, idnt.Name)
				continue
			}
			// The field is firmly considered a subcommand at this point.
			subcmd := command{
				parentNames: append(c.parentNames, c.UsgName()),
				TypeName:    idnt.Name,
				FieldName:   fieldName,
				Data:        b.getCmdClapData(idnt.Name),
			}
			// Recursively build this subcommand from it's own struct type definition.
			err := b.addChildren(&subcmd, subStrct)
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
			warn("skipping %s (looking for ident, unsure how to handle %T)\n", typeAndField, field.Type)
			continue
		}
		fieldType := basicTypeFromName(idnt.Name)
		if fieldType == -1 {
			warn("skipping %s: unsupported option or argument type '%s'\n", typeAndField, idnt.Name)
			continue
		}
		fieldDocs := parseComments(field.Doc)
		var hasOptCfgs, hasArgCfgs bool
		for i := range fieldDocs.configs {
			k := fieldDocs.configs[i].key
			if strings.HasPrefix(k, "opt") {
				hasOptCfgs = true
			}
			if strings.HasPrefix(k, "arg") {
				hasArgCfgs = true
			}
		}
		if hasOptCfgs {
			if hasArgCfgs {
				return fmt.Errorf("%s has both option and argument config values", typeAndField)
			}
			// The field is firmly considered an option at this point.
			err := c.addOption(fieldDocs, fieldName, fieldType)
			if err != nil {
				return fmt.Errorf("parsing %s field as option: %w", typeAndField, err)
			}
			continue
		}
		// The field is assumed to be an argument at this point.
		if fieldType == typBool {
			return fmt.Errorf("%s: arguments cannot be type bool", typeAndField)
		}
		c.Args = append(c.Args, argInfo{
			Data:      fieldDocs,
			FieldType: fieldType,
			FieldName: fieldName,
			name:      strings.ToLower(fieldName),
		})
	}
	c.Opts = append(c.Opts, helpOption)
	return nil
}

func basicTypeFromName(name string) basicType {
	switch name {
	case "bool":
		return typBool
	case "string":
		return typString
	default:
		return -1
	}
}

func (b *builder) findStruct(name string) *ast.StructType {
	var strct *ast.StructType
	ast.Inspect(b.pkg, func(n ast.Node) bool {
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

func (b *builder) getCmdClapData(typ string) clapData {
	var commentGrp *ast.CommentGroup
	ast.Inspect(b.pkg, func(n ast.Node) bool {
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
	c.Opts = append(c.Opts, optInfo{
		FieldType: typ,
		FieldName: fieldName,
		Long:      long,
		Short:     short,
		Data:      data,
	})
	return nil
}

func parseComments(cg *ast.CommentGroup) clapData {
	if cg == nil {
		return clapData{}
	}

	cd := clapData{
		configs: make([]clapConfig, 3),
	}

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
			cd.configs = append(cd.configs, cfg)
			lines = append(lines[:i], lines[i+1:]...)
		}
	}
	for i := range lines {
		if lines[i] == "" {
			cd.Blurb = strings.Join(lines[:i], " ")
			break
		}
	}

	// todo(steve): read in the longer description if it's there

	return cd
}
