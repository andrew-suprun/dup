package analyzer

import (
	"dup/dump"
	"dup/metas"
	"fmt"
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

type ByHashPath map[string]map[string]*MetaLink

type MetaLink struct {
	Meta *metas.FileMeta
	Link *metas.FileMeta
}

type Branch struct {
	Links         map[string]*MetaLink
	ChildBranches map[string]*Branch
}

func Analyze(origin metas.FileMetas, copy metas.FileMetas) Analysis {
	originMetas := byHashPath(origin)
	copyMetas := byHashPath(copy)
	result := Analysis{}
	for hash, copyMetaMap := range copyMetas {
		if originMetaMap, ok := originMetas[hash]; ok {
			linkMetaMaps(originMetaMap, copyMetaMap)
		}
	}

	filesToPurge(copyMetas, &result)

	fmt.Println("METAS", dump.Dump(origin))
	root := buildTree(originMetas)
	fmt.Println("TREE:", dump.Dump(root))

	filesToCopy(originMetas, &result)
	return result
}

func linkMetaMaps(originMetaMap map[string]*MetaLink, copyMetaMap map[string]*MetaLink) {
	for _, copyMetaLink := range copyMetaMap {
		var originMetaLink *MetaLink
		for _, link := range originMetaMap {
			if link.Link != nil {
				continue
			}
			if originMetaLink == nil {
				originMetaLink = link
			} else if link.Meta.Path == copyMetaLink.Meta.Path {
				originMetaLink = link
				break
			}
		}
		if originMetaLink != nil {
			originMetaLink.Link = copyMetaLink.Meta
			copyMetaLink.Link = originMetaLink.Meta
		}
	}
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

func filesToPurge(copyMeta ByHashPath, analysis *Analysis) {
	for _, copyMetaMap := range copyMeta {
		for _, copyMetaLink := range copyMetaMap {
			if copyMetaLink.Link == nil {
				analysis.Purge = append(analysis.Purge, copyMetaLink.Meta.Path)
			}
		}
	}
}

func filesToCopy(originMeta ByHashPath, analysis *Analysis) {
	for _, originMetaMap := range originMeta {
		for _, originMetaLink := range originMetaMap {
			if originMetaLink.Link == nil {
				analysis.Copy = append(analysis.Copy, originMetaLink.Meta.Path)
			}
		}
	}
}

func byHashPath(fileMetas metas.FileMetas) ByHashPath {
	result := ByHashPath{}
	for i := range fileMetas {
		meta := &fileMetas[i]
		if metaMap, ok := result[meta.Hash]; ok {
			metaMap[meta.Path] = &MetaLink{Meta: meta}
		} else {
			result[meta.Hash] = map[string]*MetaLink{meta.Path: {Meta: meta}}
		}
	}
	return result
}
