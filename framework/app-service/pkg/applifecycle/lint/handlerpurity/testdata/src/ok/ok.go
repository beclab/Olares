// Package ok exercises the "pure handler" happy path: no goroutines,
// no blocking calls, no cluster mutations. The analyzer must report
// zero diagnostics on this file.
package ok

func Handle() string {
	msg := helper()
	return msg
}

func helper() string {
	return "pure"
}
