package job

import (
	"housekeeping/internal/global"

	"go.uber.org/zap"
)

type MetadataProvider interface {
	CatIndices() CatIndice
	CatIndicesWithPattern(indexList []string) CatIndice
	CatNodes() CatNode
	CatNodesWithNodeName(nodeName string) CatNode
	CatShardsByNodeName(nodeName string) CatShard
	CatIndicesByNodeName(nodeName string) []string
}

type esMetadataProvider struct{}

func (esMetadataProvider) CatIndices() CatIndice {
	return CatIndices()
}

func (esMetadataProvider) CatIndicesWithPattern(indexList []string) CatIndice {
	return CatIndices_withPattern(indexList)
}

func (esMetadataProvider) CatNodes() CatNode {
	return CatNodes()
}

func (esMetadataProvider) CatNodesWithNodeName(nodeName string) CatNode {
	return CatNodesWithNodeName(nodeName)
}

func (esMetadataProvider) CatShardsByNodeName(nodeName string) CatShard {
	return CatShardsbyNodeName(nodeName)
}

func (esMetadataProvider) CatIndicesByNodeName(nodeName string) []string {
	return CatIndicesbyNodeName(nodeName)
}

var metadataProvider MetadataProvider = esMetadataProvider{}

func setMetadataProviderForTest(provider MetadataProvider) func() {
	prev := metadataProvider
	metadataProvider = provider
	return func() {
		metadataProvider = prev
	}
}

func ensureTestLogger() {
	if global.Logger == nil {
		global.Logger = zap.NewNop().Sugar()
	}
}
