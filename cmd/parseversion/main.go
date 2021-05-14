package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ActiveState/langtools/pkg/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

const appVersion = "0.0.7"

func main() {
	pv, err := new()
	if err != nil {
		pv.app.FatalUsage("%s\n", err)
	}

	if pv.printVersion {
		fmt.Fprintf(os.Stdout, "version %s\n", appVersion)
		os.Exit(0)
	}

	count := len(pv.args)
	if count%2 == 1 || count == 0 {
		pv.app.FatalUsage("You must pass one or more pairs of arguments, where each pair consists of a type and version string.\n")
	}

	var output []*version.Version
	for i := 0; i < count; i += 2 {
		typ := pv.args[i]
		ver := pv.args[i+1]

		var parsed *version.Version

		switch typ {
		case "generic":
			parsed, err = version.ParseGeneric(ver)
		case "semver":
			parsed, err = version.ParseSemVer(ver)
		case "perl":
			parsed, err = version.ParsePerl(ver)
		case "php":
			parsed, err = version.ParsePHP(ver)
		case "python":
			parsed, err = version.ParsePython(ver)
		case "ruby":
			parsed, err = version.ParseRuby(ver)
		default:
			pv.app.FatalUsage("Unknown version type requested: %s\n", typ)
		}

		if err != nil {
			pv.app.FatalUsage("Error parsing %s as %s: %s\n", ver, typ, err)
		}

		output = append(output, parsed)
	}

	j, err := json.Marshal(output)
	if err != nil {
		log.Fatalf("Error marshalling %+v as JSON: %s", output, err)
	}

	fmt.Println(string(j))
}

type parseversion struct {
	app          *kingpin.Application
	printVersion bool
	args         []string
}

const extraDocs = `

This command parses one or more versions and emits a JSON array containing one
or more objects describing those versions. Currently these JSON obejcts have
two keys:

  * "version" - The original string.
  * "sortable_version" - An array of strings. Each element of the array is a
    stringified decimal number. Taken as a whole, this array can be sorted
    _numerically_ against other versions of the same package.

The following version types are available:

  * semver - A version following the semver specification (https://semver.org/)
  * python - A Python PEP440 or legacy version
  * perl - A Perl module version
  * generic - Anything not covered by another type, such as C libraries, etc.
`

func new() (*parseversion, error) {
	app := kingpin.New("parseversion", "A command line tool for parsing version strings.").
		Author("ActiveState, Inc. <info@activestate.com>").
		Version(appVersion).
		UsageWriter(os.Stdout).
		UsageTemplate(kingpin.DefaultUsageTemplate + extraDocs)
	app.HelpFlag.Short('h')

	args := app.Arg(
		"type/version pairs",
		"One or more pairs of version types and versions to parse",
	).Required().Strings()

	pv := &parseversion{app: app}

	_, err := app.Parse(os.Args[1:])

	pv.args = *args

	return pv, err
}
