package analyzer

import (
	"dup/metas"
	"path/filepath"
)

type Analysis struct {
	Move []Move
	Copy []string
}

type Move struct {
	From string
	To   string
}

type Folder = metas.Folder
type Folders = metas.Folders
type File = metas.File
type Files = metas.Files

func Analyze(origin *Folder, copy *Folder, tempName string) *Analysis {
	analysis := &Analysis{}

	moveFolders(analysis, origin, copy)
	removeIdentical(origin, copy)

	return analysis
}

func moveFolders(analysis *Analysis, originFolder, copyFolder *Folder) {
	copyById := map[string]*Folder{}
	collectFoldersById(copyFolder, copyById)
	recMoveFolders(analysis, originFolder, copyById)
}

func recMoveFolders(analysis *Analysis, originFolder *Folder, copyById map[string]*Folder) {
	if originFolder.Id != "" {
		copyFolder := copyById[originFolder.Name]
		if copyFolder != nil && originFolder.Id == copyFolder.Id {
			originPath := originFolder.Path()[1:]
			analysis.Move = append(analysis.Move, Move{
				From: filepath.Join(copyFolder.Path()[1:]...),
				To:   filepath.Join(originPath...),
			})
			copyFolder.Move(originPath)
		}
	}
	for _, originSubfolder := range originFolder.Subfolders {
		recMoveFolders(analysis, originSubfolder, copyById)
	}
}

func collectFoldersById(folder *Folder, byId map[string]*Folder) {
	if folder.Id != "" {
		byId[folder.Id] = folder
	}
	for _, subfolder := range folder.Subfolders {
		collectFoldersById(subfolder, byId)
	}
}

func removeIdentical(originFolder, copyFolder *Folder) {
	for fileName, originFile := range originFolder.Files {
		copyFile := copyFolder.Files[fileName]
		if copyFile != nil && originFile.Hash == copyFile.Hash && originFile.Size == copyFile.Size && originFile.ModTime.Equal(copyFile.ModTime) {
			originFile.Remove()
			copyFile.Remove()
		}
	}
	for folderName, originSubfolder := range originFolder.Subfolders {
		copySubfolder := copyFolder.Subfolders[folderName]
		if copySubfolder != nil {
			removeIdentical(originSubfolder, copySubfolder)
		}
	}
}
