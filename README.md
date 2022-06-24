og
===

🎩 Go testing with pizazz! 👒

Go testing is great however the output is hard to scan, it takes longer to find
issues than it should and it is either too verbose, or gives no information at all.
I personally wanted more.

## Installing
You can install and use this quickly by running:

```
➜ go install github.com/tanema/og
➜ og --version
```

## Test Targeting
`og` makes it easy to address a package, file, and single test. Less typing, and
easier targeting. For instance to target a single test with `go test` you would
user:

```
go test -run TestTheTestName ./lib/...
```

To do this with `og` you can simply type:

```
og TestTheTestName
```

Some of the many ways you can address tests are:
- `og` run all tests recursively
- `og TestTheTestName TestTheOtherName` run tests by name
- `og ./lib/pack/object_test.go:20` run single test at line 20
- `og ./object_test.go` run all tests in `./object_test.go`
- `og ./object.go` run all tests in `./object_test.go` or the package if it doesnt exist
- `og ./lib/...` same as the og go test.

## Display

### Build Error Formatting

### Failure Formatting

### Test Skip Summary

### Coverage Display

## Global config
The whole point of this tool is do less typing and see pretty colors. So instead
of specifying what you want to see each time you run the command, you can define
`~/.config/og.json` and define the flags in there. You can see an example here:

```json
{
  "display": "dots",
  "split": false,
  "hide_excerpts": false,
  "hide_elapsed": false,
  "threshold": 10s,
  "no_cover": false
}
```
