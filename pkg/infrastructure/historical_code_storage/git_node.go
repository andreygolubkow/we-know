package historical_code_storage

import "os"

type GitNode struct {
	name string
	path string
	next []*FileTreeNode
}

func NewGitNode(name string, path string) *GitNode {
	return &GitNode{name: name, path: path}
}

func (g GitNode) GetName() string {
	return g.name
}

func (g GitNode) GetNext(ignoredFiles *[]string) []*FileTreeNode {
	if g.next != nil {
		return g.next
	}

	g.next = g.Discover(ignoredFiles)

	return g.next
}

func (g GitNode) Discover(ignoredFiles *[]string) []*FileTreeNode {
	fileInfo, err := os.Stat(g.path)

	if err != nil || !fileInfo.IsDir() {
		return []*FileTreeNode{}
	}

	file, err := os.Open(g.path)
	if err != nil {
		return []*FileTreeNode{}
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	fileInfos, err := file.Readdir(-1)
	if err != nil {
		return []*FileTreeNode{}
	}

	var ignored = []string{}
	if ignoredFiles != nil {
		ignored = *ignoredFiles
	}

	var nodes []*FileTreeNode
	for _, fileInfo := range fileInfos {
		if contains(ignored, fileInfo.Name()) {
			continue
		}
		var node FileTreeNode = NewGitNode(fileInfo.Name(), g.path+"/"+fileInfo.Name())
		nodes = append(nodes, &node)
	}

	return nodes
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
