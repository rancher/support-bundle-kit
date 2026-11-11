// Renovate runs `go generate ./...` after `go mod vendor` (postUpdateOptions:
// goGenerate in .github/renovate.json), so this re-applies
// patches/simulator.patch to the regenerated vendor directory.
//go:generate make simulator-vendor-patch

package main
