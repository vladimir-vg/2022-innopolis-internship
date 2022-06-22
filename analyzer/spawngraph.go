package main

import (
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

type spawngraph struct {
	root   *ssa.Function
	spawns map[*ssa.Function][]*callgraph.Edge
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

func collectSpawnFunctions(result map[*callgraph.Edge]bool, graph *callgraph.Graph, entryFunc *ssa.Function) {
	for _, edge := range graph.Nodes[entryFunc].Out {
		switch edge.Site.(type) {
		case *ssa.Go:
			result[edge] = true
		case *ssa.Call:
			collectSpawnFunctions(result, graph, edge.Callee.Func)
		case *ssa.Defer:
			collectSpawnFunctions(result, graph, edge.Callee.Func)
		default:
			panic("Unknown CallInstruction type")
		}
	}
}

func buildSpawnDAG(sgraph *spawngraph, graph *callgraph.Graph, currentFunc *ssa.Function) {
	spawnCalls := map[*callgraph.Edge]bool{}
	collectSpawnFunctions(spawnCalls, graph, currentFunc)
	spawns := make([]*callgraph.Edge, len(spawnCalls))
	i := 0
	for f := range spawnCalls {
		spawns[i] = f
		i++
	}
	sgraph.spawns[currentFunc] = spawns

	for _, edge := range spawns {
		f := edge.Callee.Func
		if _, present := sgraph.spawns[f]; !present {
			buildSpawnDAG(sgraph, graph, f)
		}
	}
}

func makeSpawnGraph(prog *ssa.Program) *spawngraph {
	graph := cha.CallGraph(prog)
	entryFunc := findFunction(prog, "main")
	sgraph := &spawngraph{
		root:   entryFunc,
		spawns: map[*ssa.Function][]*callgraph.Edge{},
	}
	buildSpawnDAG(sgraph, graph, entryFunc)
	return sgraph
}

func (sgraph *spawngraph) goroutinesRowsStream() chan goroutineRow {
	ch := make(chan goroutineRow)
	go (func() {
		// produce new rows
		for f := range sgraph.spawns {
			pos1 := f.Pos()
			pos2 := f.Prog.Fset.Position(pos1)
			ch <- goroutineRow{
				id:          f.Name(),
				packageName: f.Package().String(),
				filename:    pos2.Filename,
				line:        pos2.Line,
			}
		}
		close(ch)
	})()
	return ch
}

func (sgraph *spawngraph) goroutinesAncestryRowsStream() chan goroutineAncestryRow {
	ch := make(chan goroutineAncestryRow)
	go (func() {
		// produce new rows
		for parentF, children := range sgraph.spawns {
			for _, edge := range children {
				childF := edge.Callee.Func
				pos1 := edge.Site.Pos()
				pos2 := childF.Prog.Fset.Position(pos1)
				ch <- goroutineAncestryRow{
					parentId: parentF.Name(),
					childId:  childF.Name(),
					filename: pos2.Filename,
					line:     pos2.Line,
				}
			}
		}
		close(ch)
	})()
	return ch
}
