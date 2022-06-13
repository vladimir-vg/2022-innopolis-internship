package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"

	// "github.com/qustavo/dotsql"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
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

func loadSSA(packagePattern string) (*ssa.Program, error) {
	conf := &packages.Config{
		Dir:        "",
		Tests:      false,
		BuildFlags: []string{},
		Mode:       pkgLoadMode,
	}
	pkgPatterns := []string{packagePattern}
	loaded, err := packages.Load(conf, pkgPatterns...)
	if err != nil {
		return nil, fmt.Errorf("failed packages load: %w", err)
	}

	// fmt.Printf("%# v\n", loaded[0])

	prog, initialPkgs := ssautil.Packages(loaded, 0)

	var errorMsg bytes.Buffer
	for i, p := range initialPkgs {
		if p == nil && loaded[i].Name != "" {
			errorMsg.WriteString("failed to get SSA for pkg: ")
			errorMsg.WriteString(loaded[i].PkgPath)
			errorMsg.WriteString("\n")
		}
	}
	if errorMsg.Len() != 0 {
		return nil, errors.New(errorMsg.String())
	}

	return prog, nil
}

func main() {
	// prog, err := loadSSA("github.com/aerokube/ggr")
	prog, err := loadSSA("github.com/vladimir-vg/2022-innopolis-internship/analyzer/internal/example_package")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	prog.Build()
	sgraph := makeSpawnGraph(prog)

	fmt.Printf("#%v\n", sgraph)

	os.Remove("./foo.db")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
