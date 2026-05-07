package job

import (
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	"housekeeping/internal/structs"
)

type fakeMetadataProvider struct {
	indices CatIndice
	nodes   CatNode
	shards  CatShard
}

func (f fakeMetadataProvider) CatIndices() CatIndice {
	return f.indices
}

func (f fakeMetadataProvider) CatIndicesWithPattern(indexList []string) CatIndice {
	allowed := make(map[string]struct{}, len(indexList))
	for _, idx := range indexList {
		allowed[idx] = struct{}{}
	}

	var filtered CatIndice
	for _, idx := range f.indices {
		if _, ok := allowed[idx.Index]; ok {
			filtered = append(filtered, idx)
		}
	}
	return filtered
}

func (f fakeMetadataProvider) CatNodes() CatNode {
	return f.nodes
}

func (f fakeMetadataProvider) CatNodesWithNodeName(nodeName string) CatNode {
	var filtered CatNode
	for _, node := range f.nodes {
		if node.Name == nodeName {
			filtered = append(filtered, node)
		}
	}
	return filtered
}

func (f fakeMetadataProvider) CatShardsByNodeName(nodeName string) CatShard {
	var filtered CatShard
	for _, shard := range f.shards {
		if shard.Node == nodeName {
			filtered = append(filtered, shard)
		}
	}
	return filtered
}

func (f fakeMetadataProvider) CatIndicesByNodeName(nodeName string) []string {
	var indices []string
	for _, shard := range f.shards {
		if shard.Node == nodeName {
			indices = append(indices, shard.Index)
		}
	}
	return RemoveDuplicates(indices)
}

func makeTestIndex(name string, daysAgo int) struct {
	Health       string `json:"health"`
	Status       string `json:"status"`
	Index        string `json:"index"`
	UUID         string `json:"uuid"`
	Pri          string `json:"pri"`
	Rep          string `json:"rep"`
	DocsCount    string `json:"docs.count"`
	DocsDeleted  string `json:"docs.deleted"`
	StoreSize    string `json:"store.size"`
	PriStoreSize string `json:"pri.store.size"`
	CreationDate string `json:"creation.date"`
} {
	return struct {
		Health       string `json:"health"`
		Status       string `json:"status"`
		Index        string `json:"index"`
		UUID         string `json:"uuid"`
		Pri          string `json:"pri"`
		Rep          string `json:"rep"`
		DocsCount    string `json:"docs.count"`
		DocsDeleted  string `json:"docs.deleted"`
		StoreSize    string `json:"store.size"`
		PriStoreSize string `json:"pri.store.size"`
		CreationDate string `json:"creation.date"`
	}{
		Health:       "green",
		Status:       "open",
		Index:        name,
		UUID:         name + "-uuid",
		Pri:          "1",
		Rep:          "1",
		DocsCount:    "100",
		DocsDeleted:  "0",
		StoreSize:    "1024",
		PriStoreSize: "512",
		CreationDate: strconv.FormatInt(time.Now().AddDate(0, 0, -daysAgo).UnixMilli(), 10),
	}
}

func makeTestIndexWithStore(name string, daysAgo int, storeSizeKB string) struct {
	Health       string `json:"health"`
	Status       string `json:"status"`
	Index        string `json:"index"`
	UUID         string `json:"uuid"`
	Pri          string `json:"pri"`
	Rep          string `json:"rep"`
	DocsCount    string `json:"docs.count"`
	DocsDeleted  string `json:"docs.deleted"`
	StoreSize    string `json:"store.size"`
	PriStoreSize string `json:"pri.store.size"`
	CreationDate string `json:"creation.date"`
} {
	return struct {
		Health       string `json:"health"`
		Status       string `json:"status"`
		Index        string `json:"index"`
		UUID         string `json:"uuid"`
		Pri          string `json:"pri"`
		Rep          string `json:"rep"`
		DocsCount    string `json:"docs.count"`
		DocsDeleted  string `json:"docs.deleted"`
		StoreSize    string `json:"store.size"`
		PriStoreSize string `json:"pri.store.size"`
		CreationDate string `json:"creation.date"`
	}{
		Health:       "green",
		Status:       "open",
		Index:        name,
		UUID:         name + "-uuid",
		Pri:          "1",
		Rep:          "1",
		DocsCount:    "100",
		DocsDeleted:  "0",
		StoreSize:    storeSizeKB,
		PriStoreSize: storeSizeKB,
		CreationDate: strconv.FormatInt(time.Now().AddDate(0, 0, -daysAgo).UnixMilli(), 10),
	}
}

func sorted(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

func TestFilterSelectionPatternPrefix(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndex("logstash-api-20260220", 10),
			makeTestIndex("logstash-web-20260221", 9),
			makeTestIndex("metrics-api-20260220", 10),
			makeTestIndex(".kibana_8.0.0", 10),
		},
	})
	defer restore()

	actual := Filter_of_filter([]string{"pattern"}, []structs.Filter{
		{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
	})

	expected := []string{"logstash-api-20260220", "logstash-web-20260221"}
	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionPatternRegex(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndex("logstash-api-20260220", 10),
			makeTestIndex("logstash-api-20260221", 9),
			makeTestIndex("logstash-web-20260220", 10),
			makeTestIndex("metrics-api-20260220", 10),
		},
	})
	defer restore()

	actual := Filter_of_filter([]string{"pattern"}, []structs.Filter{
		{Filtertype: "pattern", Kind: "regex", Value: []string{`^logstash-api-\d{8}$`}},
	})

	expected := []string{"logstash-api-20260220", "logstash-api-20260221"}
	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionPatternInvalidRegex(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndex("logstash-api-20260220", 10),
			makeTestIndex("logstash-api-20260221", 9),
			makeTestIndex(".kibana_8.0.0", 10),
		},
	})
	defer restore()

	actual := Filter_of_filter([]string{"pattern"}, []structs.Filter{
		{Filtertype: "pattern", Kind: "regex", Value: []string{`(?!)invalid`}},
	})

	if len(actual) != 0 {
		t.Fatalf("unexpected matched indices for invalid regex: got %v want empty", actual)
	}
}

func TestFilterSelectionPatternSuffix(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndex("logstash-api-20260220", 10),
			makeTestIndex("metrics-api-20260220", 10),
			makeTestIndex("logstash-api-20260221", 9),
			makeTestIndex(".kibana_20260220", 10),
		},
	})
	defer restore()

	actual := Filter_of_filter([]string{"pattern"}, []structs.Filter{
		{Filtertype: "pattern", Kind: "suffix", Value: []string{"-20260220"}},
	})

	expected := []string{"logstash-api-20260220", "metrics-api-20260220"}
	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionPatternMultipleValues(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndex("logstash-api-20260220", 10),
			makeTestIndex("metrics-api-20260220", 10),
			makeTestIndex("audit-api-20260220", 10),
			makeTestIndex(".security-7", 10),
		},
	})
	defer restore()

	actual := Filter_of_filter([]string{"pattern"}, []structs.Filter{
		{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-", "metrics-"}},
	})

	expected := []string{"logstash-api-20260220", "metrics-api-20260220"}
	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionPatternExcludesSystemIndices(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndex(".kibana_8.0.0", 10),
			makeTestIndex(".security-7", 10),
			makeTestIndex("logstash-api-20260220", 10),
		},
	})
	defer restore()

	actual := Filter_of_filter([]string{"pattern"}, []structs.Filter{
		{Filtertype: "pattern", Kind: "regex", Value: []string{`.*`}},
	})

	expected := []string{"logstash-api-20260220"}
	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionAgeAndPattern(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndex("logstash-api-20260101", 40),
			makeTestIndex("logstash-api-20260220", 5),
			makeTestIndex("metrics-api-20260101", 40),
			makeTestIndex(".security-7", 40),
		},
	})
	defer restore()

	actual := Filter_of_filter([]string{"age", "pattern"}, []structs.Filter{
		{Filtertype: "age", Source: "creation_date", Direction: "older", Unit: "days", UnitCount: 30},
		{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
	})

	expected := []string{"logstash-api-20260101"}
	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionAgeRangeAndPattern(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndex("logstash-api-20260220", 20),
			makeTestIndex("logstash-api-20260221", 10),
			makeTestIndex("logstash-api-20260222", 3),
			makeTestIndex("metrics-api-20260220", 20),
		},
	})
	defer restore()

	actual := Filter_of_filter([]string{"age", "pattern"}, []structs.Filter{
		{Filtertype: "age", Source: "creation_date", Direction: "range", Unit: "days", RangeFrom: 25, RangeTo: 5},
		{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
	})

	expected := []string{"logstash-api-20260220", "logstash-api-20260221"}
	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionAgeOnly(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndex("logstash-api-20260101", 40),
			makeTestIndex("metrics-api-20260102", 35),
			makeTestIndex("logstash-api-20260220", 5),
		},
	})
	defer restore()

	actual := Filter_of_filter([]string{"age"}, []structs.Filter{
		{Filtertype: "age", Source: "creation_date", Direction: "older", Unit: "days", UnitCount: 30},
	})

	expected := []string{"logstash-api-20260101", "metrics-api-20260102"}
	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionNodeRoleAndPatternAcrossMultipleNodes(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndex("logstash-hot-20260220", 10),
			makeTestIndex("logstash-warm-a-20260220", 10),
			makeTestIndex("logstash-warm-b-20260220", 10),
			makeTestIndex("metrics-warm-20260220", 10),
		},
		nodes: CatNode{
			{Name: "hot-1", NodeRole: "h"},
			{Name: "warm-1", NodeRole: "w"},
			{Name: "warm-2", NodeRole: "w"},
		},
		shards: CatShard{
			{Index: "logstash-hot-20260220", Node: "hot-1", Shard: "0", Store: "100", Docs: "10"},
			{Index: "logstash-warm-a-20260220", Node: "warm-1", Shard: "0", Store: "100", Docs: "10"},
			{Index: "logstash-warm-b-20260220", Node: "warm-2", Shard: "0", Store: "100", Docs: "10"},
			{Index: "metrics-warm-20260220", Node: "warm-2", Shard: "1", Store: "100", Docs: "10"},
		},
	})
	defer restore()

	actual := Filters_With_node([]string{"w"}, []string{"node_role", "pattern"}, []structs.Filter{
		{Filtertype: "node_role", Value: []string{"w"}},
		{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
	})

	expected := []string{"logstash-warm-a-20260220", "logstash-warm-b-20260220"}
	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionAgeNodeRoleAndPatternAcrossMultipleNodes(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndex("logstash-hot-old-20260101", 40),
			makeTestIndex("logstash-warm-old-a-20260101", 40),
			makeTestIndex("logstash-warm-old-b-20260102", 35),
			makeTestIndex("logstash-warm-new-20260220", 5),
			makeTestIndex("metrics-warm-old-20260101", 40),
		},
		nodes: CatNode{
			{Name: "hot-1", NodeRole: "h"},
			{Name: "warm-1", NodeRole: "w"},
			{Name: "warm-2", NodeRole: "w"},
		},
		shards: CatShard{
			{Index: "logstash-hot-old-20260101", Node: "hot-1", Shard: "0", Store: "100", Docs: "10"},
			{Index: "logstash-warm-old-a-20260101", Node: "warm-1", Shard: "0", Store: "100", Docs: "10"},
			{Index: "logstash-warm-old-b-20260102", Node: "warm-2", Shard: "0", Store: "100", Docs: "10"},
			{Index: "logstash-warm-new-20260220", Node: "warm-2", Shard: "1", Store: "100", Docs: "10"},
			{Index: "metrics-warm-old-20260101", Node: "warm-1", Shard: "1", Store: "100", Docs: "10"},
		},
	})
	defer restore()

	actual := Filters_With_node([]string{"w"}, []string{"age", "node_role", "pattern"}, []structs.Filter{
		{Filtertype: "node_role", Value: []string{"w"}},
		{Filtertype: "age", Source: "creation_date", Direction: "older", Unit: "days", UnitCount: 30},
		{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
	})

	expected := []string{"logstash-warm-old-a-20260101", "logstash-warm-old-b-20260102"}
	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionPatternAndSpace(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndexWithStore("logstash-api-20260220", 30, "700000"),
			makeTestIndexWithStore("logstash-api-20260221", 20, "700000"),
			makeTestIndexWithStore("logstash-api-20260222", 10, "700000"),
			makeTestIndexWithStore("metrics-api-20260220", 30, "700000"),
		},
	})
	defer restore()

	actual := Filter_of_filter([]string{"pattern", "space"}, []structs.Filter{
		{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
		{Filtertype: "space", DiskSpace: 1},
	})

	expected := []string{"logstash-api-20260220", "logstash-api-20260221"}
	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionPatternAndWaterLevel(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndexWithStore("logstash-api-20260220", 30, "12000"),
			makeTestIndexWithStore("logstash-api-20260221", 10, "9000"),
			makeTestIndexWithStore("metrics-api-20260220", 30, "15000"),
		},
		nodes: CatNode{
			{Name: "data-1", NodeRole: "h", DiskTotal: "1gb", DiskUsedPercent: "85", DiskUsed: "0.85gb", DiskAvailable: "0.15gb"},
		},
	})
	defer restore()

	actual := Filter_of_filter([]string{"pattern", "water_level"}, []structs.Filter{
		{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
		{Filtertype: "water_level", UpperLimit: 80, LowerLimit: 79},
	})

	expected := []string{"logstash-api-20260220"}
	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionPatternAndSpaceExactThreshold(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndexWithStore("logstash-api-20260220", 30, "524288"),
			makeTestIndexWithStore("logstash-api-20260221", 20, "524288"),
			makeTestIndexWithStore("logstash-api-20260222", 10, "262144"),
		},
	})
	defer restore()

	actual := Filter_of_filter([]string{"pattern", "space"}, []structs.Filter{
		{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
		{Filtertype: "space", DiskSpace: 1},
	})

	expected := []string{"logstash-api-20260220"}
	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices at exact space threshold: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionPatternAndSpaceLargeDataset(t *testing.T) {
	ensureTestLogger()

	var indices CatIndice
	var expected []string
	for i := 1; i <= 12; i++ {
		name := "logstash-bulk-202602" + strconv.FormatInt(int64(i+9), 10)
		indices = append(indices, makeTestIndexWithStore(name, 40-i, "200000"))
		if i <= 7 {
			expected = append(expected, name)
		}
	}

	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: indices,
	})
	defer restore()

	actual := Filter_of_filter([]string{"pattern", "space"}, []structs.Filter{
		{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
		{Filtertype: "space", DiskSpace: 1},
	})

	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices for large space dataset: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionNodeRolePatternAndWaterLevel(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndexWithStore("logstash-warm-old-20260220", 30, "12000"),
			makeTestIndexWithStore("logstash-warm-new-20260221", 10, "9000"),
			makeTestIndexWithStore("metrics-warm-20260220", 30, "15000"),
			makeTestIndexWithStore("logstash-hot-20260220", 30, "12000"),
		},
		nodes: CatNode{
			{Name: "warm-1", NodeRole: "w", DiskTotal: "1gb", DiskUsedPercent: "85", DiskUsed: "0.85gb", DiskAvailable: "0.15gb"},
			{Name: "hot-1", NodeRole: "h", DiskTotal: "1gb", DiskUsedPercent: "50", DiskUsed: "0.50gb", DiskAvailable: "0.50gb"},
		},
		shards: CatShard{
			{Index: "logstash-warm-old-20260220", Node: "warm-1", Shard: "0", Store: "12000", Docs: "10"},
			{Index: "logstash-warm-new-20260221", Node: "warm-1", Shard: "1", Store: "9000", Docs: "10"},
			{Index: "metrics-warm-20260220", Node: "warm-1", Shard: "2", Store: "15000", Docs: "10"},
			{Index: "logstash-hot-20260220", Node: "hot-1", Shard: "0", Store: "12000", Docs: "10"},
		},
	})
	defer restore()

	actual := Filters_With_node([]string{"w"}, []string{"node_role", "pattern", "water_level"}, []structs.Filter{
		{Filtertype: "node_role", Value: []string{"w"}},
		{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
		{Filtertype: "water_level", UpperLimit: 80, LowerLimit: 79},
	})

	expected := []string{"logstash-warm-old-20260220"}
	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionPatternAndWaterLevelExactThreshold(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndexWithStore("logstash-api-20260220", 30, "12000"),
			makeTestIndexWithStore("logstash-api-20260221", 10, "9000"),
			makeTestIndexWithStore("metrics-api-20260220", 30, "15000"),
		},
		nodes: CatNode{
			{Name: "data-1", NodeRole: "h", DiskTotal: "1gb", DiskUsedPercent: "80", DiskUsed: "0.80gb", DiskAvailable: "0.20gb"},
		},
	})
	defer restore()

	actual := Filter_of_filter([]string{"pattern", "water_level"}, []structs.Filter{
		{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
		{Filtertype: "water_level", UpperLimit: 80, LowerLimit: 79},
	})

	expected := []string{"logstash-api-20260220"}
	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices at exact water level threshold: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionPatternAndWaterLevelLargeDataset(t *testing.T) {
	ensureTestLogger()

	var indices CatIndice
	var expected []string
	for i := 1; i <= 12; i++ {
		name := "logstash-bulk-202603" + strconv.FormatInt(int64(i+9), 10)
		indices = append(indices, makeTestIndexWithStore(name, 50-i, "2000"))
		if i <= 6 {
			expected = append(expected, name)
		}
	}

	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: indices,
		nodes: CatNode{
			{Name: "data-1", NodeRole: "h", DiskTotal: "1gb", DiskUsedPercent: "85", DiskUsed: "0.85gb", DiskAvailable: "0.15gb"},
		},
	})
	defer restore()

	actual := Filter_of_filter([]string{"pattern", "water_level"}, []structs.Filter{
		{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
		{Filtertype: "water_level", UpperLimit: 80, LowerLimit: 79},
	})

	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices for large water level dataset: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionNodeRolePatternAndWaterLevelAcrossMultipleNodes(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndexWithStore("logstash-warm-old-a-20260220", 30, "12000"),
			makeTestIndexWithStore("logstash-warm-new-a-20260221", 10, "9000"),
			makeTestIndexWithStore("logstash-warm-old-b-20260218", 35, "11000"),
			makeTestIndexWithStore("logstash-warm-new-b-20260222", 5, "8000"),
			makeTestIndexWithStore("metrics-warm-20260220", 30, "15000"),
		},
		nodes: CatNode{
			{Name: "warm-1", NodeRole: "w", DiskTotal: "1gb", DiskUsedPercent: "85", DiskUsed: "0.85gb", DiskAvailable: "0.15gb"},
			{Name: "warm-2", NodeRole: "w", DiskTotal: "1gb", DiskUsedPercent: "86", DiskUsed: "0.86gb", DiskAvailable: "0.14gb"},
		},
		shards: CatShard{
			{Index: "logstash-warm-old-a-20260220", Node: "warm-1", Shard: "0", Store: "12000", Docs: "10"},
			{Index: "logstash-warm-new-a-20260221", Node: "warm-1", Shard: "1", Store: "9000", Docs: "10"},
			{Index: "logstash-warm-old-b-20260218", Node: "warm-2", Shard: "0", Store: "11000", Docs: "10"},
			{Index: "logstash-warm-new-b-20260222", Node: "warm-2", Shard: "1", Store: "8000", Docs: "10"},
			{Index: "metrics-warm-20260220", Node: "warm-2", Shard: "2", Store: "15000", Docs: "10"},
		},
	})
	defer restore()

	actual := Filters_With_node([]string{"w"}, []string{"node_role", "pattern", "water_level"}, []structs.Filter{
		{Filtertype: "node_role", Value: []string{"w"}},
		{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
		{Filtertype: "water_level", UpperLimit: 80, LowerLimit: 79},
	})

	expected := []string{"logstash-warm-old-a-20260220", "logstash-warm-old-b-20260218"}
	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionNodeRolePatternAndSpace(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndexWithStore("logstash-warm-old-20260220", 30, "700000"),
			makeTestIndexWithStore("logstash-warm-new-20260221", 10, "700000"),
			makeTestIndexWithStore("logstash-hot-20260222", 5, "700000"),
			makeTestIndexWithStore("metrics-warm-20260220", 30, "700000"),
		},
		nodes: CatNode{
			{Name: "warm-1", NodeRole: "w"},
			{Name: "hot-1", NodeRole: "h"},
		},
		shards: CatShard{
			{Index: "logstash-warm-old-20260220", Node: "warm-1", Shard: "0", Store: "700000", Docs: "10"},
			{Index: "logstash-warm-new-20260221", Node: "warm-1", Shard: "1", Store: "700000", Docs: "10"},
			{Index: "logstash-hot-20260222", Node: "hot-1", Shard: "0", Store: "700000", Docs: "10"},
			{Index: "metrics-warm-20260220", Node: "warm-1", Shard: "2", Store: "700000", Docs: "10"},
		},
	})
	defer restore()

	actual := Filters_With_node([]string{"w"}, []string{"node_role", "pattern", "space"}, []structs.Filter{
		{Filtertype: "node_role", Value: []string{"w"}},
		{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
		{Filtertype: "space", DiskSpace: 1},
	})

	expected := []string{"logstash-warm-old-20260220"}
	if !reflect.DeepEqual(sorted(actual), sorted(expected)) {
		t.Fatalf("unexpected matched indices: got %v want %v", actual, expected)
	}
}

func TestFilterSelectionNodeRolePatternAndWaterLevelBelowThreshold(t *testing.T) {
	ensureTestLogger()
	restore := setMetadataProviderForTest(fakeMetadataProvider{
		indices: CatIndice{
			makeTestIndexWithStore("logstash-warm-old-20260220", 30, "12000"),
			makeTestIndexWithStore("logstash-warm-new-20260221", 10, "9000"),
		},
		nodes: CatNode{
			{Name: "warm-1", NodeRole: "w", DiskTotal: "1gb", DiskUsedPercent: "60", DiskUsed: "0.60gb", DiskAvailable: "0.40gb"},
		},
		shards: CatShard{
			{Index: "logstash-warm-old-20260220", Node: "warm-1", Shard: "0", Store: "12000", Docs: "10"},
			{Index: "logstash-warm-new-20260221", Node: "warm-1", Shard: "1", Store: "9000", Docs: "10"},
		},
	})
	defer restore()

	actual := Filters_With_node([]string{"w"}, []string{"node_role", "pattern", "water_level"}, []structs.Filter{
		{Filtertype: "node_role", Value: []string{"w"}},
		{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
		{Filtertype: "water_level", UpperLimit: 80, LowerLimit: 79},
	})

	if len(actual) != 0 {
		t.Fatalf("unexpected matched indices below water level threshold: got %v want empty", actual)
	}
}
