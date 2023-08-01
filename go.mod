module github.com/frankban/quicktest

require (
	github.com/google/go-cmp v0.5.9
	github.com/kr/pretty v0.3.1
)

require (
	github.com/kr/text v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
)

// We do actually support go 1.14, but until go 1.21
// we can't have any generics code even if guarded
// by a build tag.
go 1.18
