# What Is This

The langtools repo contains packages and tools that we at ActiveState have
developed as part of the [ActiveState
Platform](https://platform.activestate.com/). The platform provides automated
language builds, where you can pick a language core and a set of packages to
be built on a variety of platforms. Since building the platform requires us to
understand a number of language package ecosystems, we are building tools for
working with these ecosystems.

## Version Parsing

This repo contains a Go package for version parsing,
`github.com/ActiveState/langtools/pkg/version`:

```go
package main

import (
	"fmt"
	"log"

	"github.com/ActiveState/langtools/pkg/version"
)

func main() {
	v, err := version.ParseGeneric("1.2")
	if err != nil {
		log.Fatalf("Could not parse 1.2 as a generic version: %s", err)
	}
	fmt.Printf("Parsed as %v\n", v.Decimal)
	// Prints:
	// Parsed as [1 2]
}
```

## Name Normalization

Some language ecosystems have a concept of name normalization for package
names. This repo contains a Go package for name normalization,
`github.com/ActiveState/langtools/pkg/name`:

```go
package main

import (
	"fmt"

	"github.com/ActiveState/langtools/pkg/name"
)

func main() {
	norm := name.NormalizePython("backports.functools_lru_cache")
	fmt.Printf("Normalized as %s\n", norm)
	// Prints:
	// Normalized as backports-functools-lru-cache
}
```

### `parseversion` Command Line Tool

This repository also contains the code for a `parseversion` CLI tool. You can
install this by running `go get
github.com/ActiveState/langtools/cmd/parseversion`. Run `parseversion --help`
for details on this tool.

## Build Status

[![CircleCI](https://circleci.com/gh/ActiveState/langtools.svg?style=svg)](https://circleci.com/gh/ActiveState/langtools)

## To Create a New Release
* Tag master: `git tag v0.0.6`
* Push the tag: `git push origin v0.0.6`
* Go to the [releases page](https://github.com/ActiveState/langtools/releases)
* Click on the name of the new release, `v0.0.6` in this example
* Click the `Edit Release` button on the top right
* Enter in the release notes, which are the same as you added to `Changes.md`
  * Remove any manual line breaks as they don't look good
* Click the `Publish Release` button on the bottom

## Authors

This library was created by:

* Sean Fitzgerald
* Jason Palmer
* Dave Rolsky \<autarch@urth.org\>
* Tyler Santerre
* Stephen Reichling

## Copyright

Copyright (c) 2020-2024, ActiveState Software.
All rights reserved.

## License

This software is licensed under the BSD 3-Clause License.
