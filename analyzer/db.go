package main

type goroutineRow struct {
	id          string
	packageName string
	filename    string
	line        int
}

type goroutineAncestryRow struct {
	id       string
	parentId string
	childId  string
	filename string
	line     int
}

type fileRow struct {
	filename string
	content  []byte
}
