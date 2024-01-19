package main

import (
	"log"
	"os"
	hs "we-know/pkg/infrastructure/historical_code_storage"
)

const (
	WrongArgumentsException = "Nothing to clone"
)

func main() {
	repoUrl, repoBranch, repoDir := readArguments()

	codeStorage, err := hs.NewGitStorage(repoUrl, repoBranch, repoDir)

	if err != nil {
		log.Fatal(err)
	}

	err = codeStorage.SetUp()

	if err != nil {
		log.Fatal(err)
		return
	}

	var rootPtr = codeStorage.GetRootNode()
	if rootPtr == nil {
		log.Fatal("Root node is nil")
		return
	}
	var root = *rootPtr
	log.Print(root.GetName())

	var ignoreList = []string{".git", ".idea", ".github"}
	Walk(rootPtr, func(node *hs.FileTreeNode) {
		log.Print((*node).GetName())
	}, &ignoreList)
}

type treeCallback func(node *hs.FileTreeNode)

func Walk(root *hs.FileTreeNode, callback treeCallback, ignoredFiles *[]string) {
	if root == nil {
		return
	}
	var r = *root
	nextNodes := r.GetNext(ignoredFiles)
	for _, node := range nextNodes {
		callback(node)
		Walk(node, callback, ignoredFiles)
	}
}

/*
 */
func readArguments() (repositoryUrl string, repoBranch string, repositoryDir string) {
	if len(os.Args) != 3 {
		log.Fatal(WrongArgumentsException)
	}
	repositoryUrl = os.Args[1]
	branch := os.Args[2]
	workingDir, _ := os.Getwd()
	tmpDir := workingDir + "/tmp"

	info, err := os.Stat(tmpDir)

	if os.IsNotExist(err) || !info.IsDir() {
		err := os.MkdirAll(tmpDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	return repositoryUrl, branch, tmpDir
}
