<!-- (C) Copyright 2023 Hewlett Packard Enterprise Development LP -->

# Ruleguard rules to verify GLCP Go style guide compliance

This repository contains [ruleguard](https://github.com/quasilyte/go-ruleguard) rules that verify as much as possible
compliance with the [GLCP Go Style guide](https://developer.greenlake.hpe.com/docs/greenlake/standards/language/go_style_guide). Since the guide is opinionated, so are the rules. The rules are incomplete in that they will not verify every aspect of the guide and should not be used as the sole alternative to reading and understanding the aims of the guide. Instead, they should be viewed as an additional check of some easy items and not as a replacement for human review of the code. They are designed to enable more focussed human review of the code by automating trivial style compliance.

## Using the rules

The best way of using these rules is to run them as part of the [gocritic](https://github.com/go-critic/go-critic) linter called from within the [golangci-lint](https://github.com/golangci/golangci-lint) framework.

## Referencing the rules in a repo

The preferred way of using these rules is to leverage the use of Go module support and use the rules defined in this repo.  

Create a directory ```rules``` and add a file called `rules.go`

```
// (C) Copyright 2023 Hewlett Packard Enterprise Development LP

//go:build ruleguard
// +build ruleguard

// Package rules imports the style guide ruleguard bundle.
package rules

import (
	glcp "github.com/hpe-hcss/ruleguard-go-guide"
	"github.com/quasilyte/go-ruleguard/dsl"
)

func init() {
	dsl.ImportRules("glcp", glcp.Bundle)
}
```

and then run ` go mod tidy`. This will add a reference to this repository to access the rules and will track changes over time that may occur to the rules to track any changes made to the style guide. The Go build system will not use this source during usual build because of the build tags. Updates can be managed using standard `dependabot` or through use of `go get -u ...`.

### Enabling gocritic to use ruleguard

Turing on ruleguard support in gocritic requires a few changes to the golangci-lint configuration file. Under the `linters-settings` add or create a section for `gocritic` and enable `ruleguard`. Then instruct ruleguard where to find the rules to apply.

```
linters-settings:
  gocritic:
    enabled-checks:
      - ruleguard
    settings:
      ruleguard:
        rules: '${configDir}/rules/rules.go'
        failOn: dsl
```

## Verifying that the rules are working

Write some code inconsistent with the style.

<table>
<thead><tr><th>Bad</th><th>Rules in action</th></tr></thead>
<tbody>
<tr><td>

```go
func main() {
    var x = 0
}
```

</td><td>

```go
func main() {
    var x = 0 // ruleguard: var used for assignment; use 'x := 0' (gocritic)
}
```

</td></tr>
<tr><td>

```go
type smap struct {
    sync.Mutex
}
```

</td><td>

```go
type smap struct {
    sync.Mutex // ruleguard: do not embed sync.Mutex (gocritic)
}
```

</td></tr>
</tbody></table>

## Suppressing lint warnings  

The suppression of a style-guide lint error should be the path of last resort. Clear, in code, comments must be provided as to why it is appropriate to deviate from the guide. The actual suppression of the rule can be accomplished by placing the comment `//nolint: gocritic` next to, or above, the line that would ordinarily be flagged by the linter.

## Running ruleguard manually 

Install ruleguard
```
go install -v github.com/quasilyte/go-ruleguard/cmd/ruleguard@latest
```
run it with the rules
```
ruleguard -rules rules/rules.go *.go
```
 # Enhancements

 Currently none of the rules contain a Suggestion, doing so would enable automatic style compliance through the direct editing of the source files when golangci-lint (or ruleguard) are run with the `-fix` option. This would require up-front suppression (via //nolint) on lines that would otherwise be re-written. 