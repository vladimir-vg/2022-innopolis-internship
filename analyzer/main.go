package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/qustavo/dotsql"

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

	os.Remove("../dev.db")
	db, err := sql.Open("sqlite3", "../dev.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dot, _ := dotsql.LoadFromFile("../queries.sql")
	// _, err :=
	dot.Exec(db, "initialize")
	if err != nil {
		log.Fatal(err)
	}

	for row := range sgraph.goroutinesRowsStream() {
		_, err := dot.Exec(
			db, "insert-goroutine",
			row.id, row.packageName, row.filename, row.line,
		)
		if err != nil {
			log.Fatal(err)
		}
	}
	for row := range sgraph.goroutinesAncestryRowsStream() {
		_, err := dot.Exec(
			db, "insert-spawn",
			row.parentId, row.childId, row.filename, row.line,
		)
		if err != nil {
			log.Fatal(err)
		}
	}
}
