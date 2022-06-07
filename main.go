package main

import (
	"fmt"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa/ssautil"
)

const pkgLoadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedCompiledGoFiles |
	packages.NeedImports |
	packages.NeedDeps |
	packages.NeedExportsFile |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo |
	packages.NeedTypesSizes |
	packages.NeedModule

func main() {
	conf := &packages.Config{
		Dir:        "",
		Tests:      false,
		BuildFlags: []string{},
		Mode:       pkgLoadMode,
	}
	pkgPatterns := []string{"github.com/aerokube/ggr"}
	loaded, err := packages.Load(conf, pkgPatterns...)
	if err != nil {
		fmt.Errorf("failed packages load: %w", err)
	}
	fmt.Printf("%#v\n", loaded[0])
	ssautil.Packages(loaded, 0)
}
