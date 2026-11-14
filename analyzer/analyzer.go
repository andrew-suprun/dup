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

type ByHashPath map[string]map[string]MetaLink

type MetaLink struct {
	Meta *metas.FileMeta
	Link *metas.FileMeta
}

func Analyze(origin metas.FileMetas, copy metas.FileMetas) Analysis {
	originMetas := byHashPath(origin)
	copyMetas := byHashPath(copy)
	result := Analysis{}
	for hash, metaMap := range copyMetas {
		if _, ok := originMetas[hash]; !ok {
			for path := range metaMap {
				fmt.Println("purge", path)
				result.Purge = append(result.Purge, path)
			}
		}
	}

	return result
}

func byHashPath(fileMetas metas.FileMetas) ByHashPath {
	result := ByHashPath{}
	for i := range fileMetas {
		meta := &fileMetas[i]
		if metaMap, ok := result[meta.Hash]; ok {
			metaMap[meta.Path] = MetaLink{Meta: meta}
		} else {
			result[meta.Hash] = map[string]MetaLink{meta.Path: {Meta: meta}}
		}
	}
	return result
}
