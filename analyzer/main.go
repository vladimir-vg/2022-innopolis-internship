package main

import (
	"bytes"
	"errors"
	"fmt"

	// "github.com/kr/pretty"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
	// "golang.org/x/tools/go/callgraph"
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

func findFunction(prog *ssa.Program, entryFunctionName string) *ssa.Function {
	// find the entry function
	// fmt.Printf("func name: %#v\n", ssautil.AllFunctions(prog))
	for f := range ssautil.AllFunctions(prog) {
		if f.Name() == entryFunctionName && f.Synthetic == "" {
			return f
		}
	}
	return nil
}

type spawngraph struct {
	root   *ssa.Function
	spawns map[*ssa.Function][]*ssa.Function
}

func collectSpawnFunctions(result map[*ssa.Function]bool, graph *callgraph.Graph, entryFunc *ssa.Function) {
	for _, edge := range graph.Nodes[entryFunc].Out {
		switch edge.Site.(type) {
		case *ssa.Go:
			result[edge.Callee.Func] = true
		case *ssa.Call:
			collectSpawnFunctions(result, graph, edge.Callee.Func)
		case *ssa.Defer:
			collectSpawnFunctions(result, graph, edge.Callee.Func)
		default:
			panic("Unknown CallInstruction type")
		}
	}
}

func buildSpawnGraph(sgraph *spawngraph, graph *callgraph.Graph, entryFunc *ssa.Function) {
	spawnCalls := map[*ssa.Function]bool{}
	collectSpawnFunctions(spawnCalls, graph, entryFunc)
	spawns := make([]*ssa.Function, len(spawnCalls))
	i := 0
	for f := range spawnCalls {
		spawns[i] = f
		i++
	}
	sgraph.spawns[entryFunc] = spawns

	for _, f := range spawns {
		if _, present := sgraph.spawns[f]; !present {
			buildSpawnGraph(sgraph, graph, f)
		}
	}
}

func main() {
	// prog, err := loadSSA("github.com/aerokube/ggr")
	prog, err := loadSSA("github.com/vladimir-vg/2022-innopolis-internship/analyzer/internal/example_package")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	prog.Build()
	mainFunc := findFunction(prog, "main")
	graph := cha.CallGraph(prog)
	// result := map[*ssa.Function]bool{}
	sgraph := &spawngraph{
		root:   mainFunc,
		spawns: map[*ssa.Function][]*ssa.Function{},
	}
	buildSpawnGraph(sgraph, graph, sgraph.root)
	// collectSpawnFunctions(result, graph, mainFunc)
	// buildSpawnGraph(graph, mainFunc)
	// funcs := []*ssa.Function{mainFunc}
	// fmt.Printf("#%v\n", funcs)
	// result := rta.Analyze(funcs, true)

	// mainFunc = prog.FuncValue()
	// graph := callgraph.New(mainFunc)

	// fmt.Printf("#%v\n", graph.Root)
	fmt.Printf("#%v\n", sgraph)

}
