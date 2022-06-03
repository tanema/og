og
===

🎩 Go testing with pizazz! 👒

**Features**
- Multiple different displays
- Customizable
- Easy test addressing
  - `og` run all tests recursively
  - `og TestTheTestName TestTheOtherName` run tests by name
  - `og ./lib/pack/object_test.go:20` run single test at line 20
  - `og ./object_test.go` run all tests in `./object_test.go`
  - `og ./object.go` run all tests in `./object_test.go` or the package if it doesnt exist
  - `og ./lib/...` same as the og go test.
