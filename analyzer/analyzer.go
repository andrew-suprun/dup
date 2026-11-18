package analyzer

import (
	"dup/metas"
	"path/filepath"
	"strings"
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

type MetaLink struct {
	Meta *metas.File
	Link *metas.File
}

type Branch struct {
	Links         map[string]*MetaLink
	ChildBranches map[string]*Branch
}

func Analyze(origin metas.Archive, copy metas.Archive, tempName string) Analysis {
	analysis := Analysis{}

	originLinks := make([]int, len(origin.Files))
	copyLinks := make([]int, len(copy.Files))

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
			originLinks[originIdx] = copyIdx
			copyLinks[copyIdx] = originIdx
		} else {
			copyLinks[copyIdx] = -1
		}
	}

	for copyIdx := range copy.Files {
		if copyLinks[copyIdx] == -1 {
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

func buildTree(metas ByHashPath) *Branch {
	root := &Branch{
		Links:         map[string]*MetaLink{},
		ChildBranches: map[string]*Branch{},
	}
	for _, metaMap := range metas {
		for _, metaLink := range metaMap {
			branch := root
			components := strings.Split(metaLink.Meta.Path, "/")
			for _, component := range components[:len(components)-1] {
				if childBranch, ok := branch.ChildBranches[component]; ok {
					branch = childBranch
				} else {
					childBranch := &Branch{
						Links:         map[string]*MetaLink{},
						ChildBranches: map[string]*Branch{},
					}
					branch.ChildBranches[component] = childBranch
					branch = childBranch
				}
			}
			branch.Links[components[len(components)-1]] = metaLink
		}
	}

	return root
}
