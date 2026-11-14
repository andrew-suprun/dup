package analyzer

import (
	"dup/metas"
	"fmt"
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

func Analyze(origin metas.FileMetas, copy metas.FileMetas) Analysis {
	originMetas := byHashPath(origin)
	copyMetas := byHashPath(copy)
	result := Analysis{}
	for hash, copyMetaMap := range copyMetas {
		if originMetaMap, ok := originMetas[hash]; ok {
			linkMetaMaps(originMetaMap, copyMetaMap)
		} else {
			for path := range copyMetaMap {
				fmt.Println("purge", path)
				result.Purge = append(result.Purge, path)
			}
		}
	}

	filesToCopy(originMetas, &result)

	return result
}

func linkMetaMaps(originMetaMap map[string]*MetaLink, copyMetaMap map[string]*MetaLink) {
	for _, copyMetaLink := range copyMetaMap {
		var originMetaLink *MetaLink
		for _, link := range originMetaMap {
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

func filesToCopy(originMeta ByHashPath, analysis *Analysis) {
	for _, originMetaMap := range originMeta {
		for _, originMetaLink := range originMetaMap {
			if originMetaLink.Link == nil {
				analysis.Copy = append(analysis.Copy, originMetaLink.Meta.Path)
				fmt.Println("copy:", originMetaLink.Meta.Path)
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
