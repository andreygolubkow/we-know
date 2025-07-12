package walker

import (
	"os"
	hs "we-know/pkg/infrastructure/historical_code_storage"
	"we-know/pkg/infrastructure/user"
)

// TreeCallback is a function type that is called for each node in the file tree
type TreeCallback func(node *hs.FileTreeNode, fullPath string)

// Walk traverses the file tree and calls the callback function for each node
func Walk(root *hs.FileTreeNode, callback TreeCallback, pathBase string, ignoredFiles *[]string) {
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

// Crawl traverses the file tree and populates the FileEditorsStorage with editor information for each file
// If userMapping is provided, it will be used to map user IDs to display names
func Crawl(root *hs.FileTreeNode, codeStorage hs.HistoricalCodeStorage, storage *hs.FileEditorsStorage, pathBase string, ignoredFiles *[]string, userMapping *user.UserMapping) {
	Walk(root, func(node *hs.FileTreeNode, path string) {
		editors, errorMsg := codeStorage.GetEditorsByFile(path)

		// If we have editors and user mapping, apply the mapping
		if editors != nil && userMapping != nil {
			mappedEditors := make(map[string]int)
			for userID, lines := range *editors {
				displayName := userMapping.GetDisplayName(userID)
				mappedEditors[displayName] += lines
			}
			mappedEditorsPtr := &mappedEditors
			storage.SetFileEditors(path, mappedEditorsPtr, errorMsg)
		} else {
			storage.SetFileEditors(path, editors, errorMsg)
		}
	}, pathBase, ignoredFiles)
}
