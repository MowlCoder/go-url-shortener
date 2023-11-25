// staticlint is tool for analyzing your code for potential errors and wrong constructions.
// This tools combines different analyzing tools.
//
// # Default go analyzers
//   - printf: checks consistency of Printf format strings and arguments
//   - shadow: checks for shadowed variables
//   - structtag: check that struct field tags conform to reflect.StructTag.Get
//   - fieldalignment: find structs that would use less memory if their fields were sorted
//
// # All staticcheck analyzers of SA group
//
// The SA category of checks, codenamed staticcheck, contains all checks that are concerned with the correctness of code.
// More here - https://staticcheck.dev/docs/checks/#SA
//
// # Simple checks
//   - S1012: replace time.Now().Sub(x) with time.Since(x)
//   - S1024: replace x.Sub(time.Now()) with time.Until(x)
//   - S1025: don’t use fmt.Sprintf("%s", x) unnecessarily
//   - S1011: use a single append to concatenate two slices
//
// # Style checks
//   - ST1006: checks for poorly chosen receiver name
//   - ST1017: don’t use Yoda conditions
//   - ST1023: redundant type in variable declaration
//   - ST1015: a switch’s default case should be the first or last case
//
// # Go-Critic analyzer
//
// Highly extensible Go source code linter providing checks currently missing from other linters.
// More here - https://github.com/go-critic/go-critic
//
// # Body closed analyzer
//
// bodyclose is a static analysis tool which checks whether res.Body is correctly closed.
// More here - https://github.com/timakin/bodyclose
//
// # osexitmain analyzer
//
// # Checks for usage of os.Exit in main function of main package
//
// # Usage
//
// To use linter you should open preferred terminal and write these commands:
//   - go install github.com/MowlCoder/go-url-shortener/cmd/staticlint
//   - staticlint ./... (to scan all files in project)
package main
