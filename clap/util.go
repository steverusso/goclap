package clap

import (
	"flag"
	"fmt"
	"os"
)

func ParseEnv(v flag.Value, name string) (bool, error) {
	s, ok := os.LookupEnv(name)
	if !ok {
		return false, nil
	}
	if err := v.Set(s); err != nil {
		return true, fmt.Errorf(`invalid value "%s" for env var "%s": %w`, s, name, err)
	}
	return true, nil
}
