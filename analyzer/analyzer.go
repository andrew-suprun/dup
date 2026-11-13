package analyzer

import (
	"fmt"
	"time"
)

type FileMeta struct {
	Path    string
	Size    int64
	ModTime time.Time
	Hash    string
}

type FileMetas []FileMeta

type Analysis struct {
	Purge []string
	Move  []Move
	Copy  []string
}

type Move struct {
	From string
	To   string
}

func Analyze(origin FileMetas, copy FileMetas) Analysis {
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

type ByHashPath map[string]map[string]*FileMeta

func byHashPath(metas FileMetas) ByHashPath {
	result := ByHashPath{}
	for i := range metas {
		meta := &metas[i]
		if metaMap, ok := result[meta.Hash]; ok {
			metaMap[meta.Path] = meta
		} else {
			result[meta.Hash] = map[string]*FileMeta{
				meta.Path: meta,
			}
		}
	}
	return result
}
