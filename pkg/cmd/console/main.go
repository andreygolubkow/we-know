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
	var pathBase = ""
	Walk(rootPtr, func(node *hs.FileTreeNode, path string) {
		//realPath := repoDir + string(os.PathSeparator) + path
		blame, _ := codeStorage.GetEditorsByFile(path)
		log.Print(blame)
	}, pathBase, &ignoreList)
}

type treeCallback func(node *hs.FileTreeNode, fullPath string)

func Walk(root *hs.FileTreeNode, callback treeCallback, pathBase string, ignoredFiles *[]string) {
	if root == nil {
		return
	}
	var r = *root
	nextNodes := r.GetNext(ignoredFiles)
	for _, node := range nextNodes {
		var path = r.GetName()
		if len(pathBase) > 0 {
			path = pathBase + string(os.PathSeparator) + r.GetName()
		}
		var callbackPath = (*node).GetName()
		if len(path) > 0 {
			callbackPath = path + string(os.PathSeparator) + (*node).GetName()
		}
		callback(node, callbackPath)
		Walk(node, callback, path, ignoredFiles)
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
