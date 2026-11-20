package analyzer

import (
	"dup/metas"
	"path/filepath"
)

type Analysis struct {
	Purge []string
	Move  []Move
	Copy  []string
}

type Move struct {
	From string
	To   string
}

type FolderLink struct {
	Meta *metas.Folder
	Link *metas.Folder
}

type FileLink struct {
	Meta *metas.File
	Link *metas.File
}

type Branch struct {
	Links         map[string]*FileLink
	ChildBranches map[string]*Branch
}

func Analyze(origin *metas.Folder, copy *metas.Folder, tempName string) Analysis {
	analysis := Analysis{}

	// TODO link files

	for copyIdx := range copy.Files {
		copyMeta := &copy.Files[copyIdx]
		originIdx := -1
		for idx := range origin.Files {
			originMeta := &origin.Files[originIdx]
			if originMeta.Hash == copyMeta.Hash {
				if originMeta.Path == copyMeta.Path || originIdx == -1 {
					originIdx = idx
				}
			}
		}
		if originIdx != -1 {
			originFileLinks[originIdx] = copyIdx
			copyFileLinks[copyIdx] = originIdx
		} else {
			copyFileLinks[copyIdx] = -1
		}
	}

	for copyIdx := range copy.Files {
		if copyFileLinks[copyIdx] == -1 {
			copyMeta := &copy.Files[copyIdx]
			analysis.Purge = append(analysis.Purge, copyMeta.Path)
			copyMeta.Path = ""
			copyMeta.Hash = ""
		}
	}

	for originFolderIdx := range origin.Foldes {
		originFolderMeta := &origin.Foldes[originFolderIdx]

		for copyFolderIdx := range copy.Foldes {
			copyFolderMeta := &copy.Foldes[copyFolderIdx]
			if originFolderMeta.Id != copyFolderMeta.Id {
				analysis.Move = append(analysis.Move, Move{
					From: copyFolderMeta.Name,
					To:   filepath.Join(copyFolderMeta.Name, tempName)})

			}
		}
	}

	return analysis
}
