package job

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"housekeeping/internal/structs"
)

type goldenCase struct {
	name     string
	file     string
	provider fakeMetadataProvider
	run      func() []string
}

func TestFilterSelectionGolden(t *testing.T) {
	ensureTestLogger()

	cases := []goldenCase{
		{
			name: "pattern_prefix",
			file: "pattern_prefix.golden.json",
			provider: fakeMetadataProvider{
				indices: CatIndice{
					makeTestIndex("logstash-api-20260220", 10),
					makeTestIndex("logstash-web-20260221", 9),
					makeTestIndex("metrics-api-20260220", 10),
					makeTestIndex(".kibana_8.0.0", 10),
				},
			},
			run: func() []string {
				return Filter_of_filter([]string{"pattern"}, []structs.Filter{
					{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
				})
			},
		},
		{
			name: "age_node_role_pattern",
			file: "age_node_role_pattern.golden.json",
			provider: fakeMetadataProvider{
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
			},
			run: func() []string {
				return Filters_With_node([]string{"w"}, []string{"age", "node_role", "pattern"}, []structs.Filter{
					{Filtertype: "node_role", Value: []string{"w"}},
					{Filtertype: "age", Source: "creation_date", Direction: "older", Unit: "days", UnitCount: 30},
					{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
				})
			},
		},
		{
			name: "pattern_space",
			file: "pattern_space.golden.json",
			provider: fakeMetadataProvider{
				indices: CatIndice{
					makeTestIndexWithStore("logstash-api-20260220", 30, "700000"),
					makeTestIndexWithStore("logstash-api-20260221", 20, "700000"),
					makeTestIndexWithStore("logstash-api-20260222", 10, "700000"),
					makeTestIndexWithStore("metrics-api-20260220", 30, "700000"),
				},
			},
			run: func() []string {
				return Filter_of_filter([]string{"pattern", "space"}, []structs.Filter{
					{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
					{Filtertype: "space", DiskSpace: 1},
				})
			},
		},
		{
			name: "pattern_space_large_dataset",
			file: "pattern_space_large_dataset.golden.json",
			provider: fakeMetadataProvider{
				indices: func() CatIndice {
					var indices CatIndice
					for i := 1; i <= 12; i++ {
						name := "logstash-bulk-202602" + strconv.FormatInt(int64(i+9), 10)
						indices = append(indices, makeTestIndexWithStore(name, 40-i, "200000"))
					}
					return indices
				}(),
			},
			run: func() []string {
				return Filter_of_filter([]string{"pattern", "space"}, []structs.Filter{
					{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
					{Filtertype: "space", DiskSpace: 1},
				})
			},
		},
		{
			name: "pattern_water_level_large_dataset",
			file: "pattern_water_level_large_dataset.golden.json",
			provider: fakeMetadataProvider{
				indices: func() CatIndice {
					var indices CatIndice
					for i := 1; i <= 12; i++ {
						name := "logstash-bulk-202603" + strconv.FormatInt(int64(i+9), 10)
						indices = append(indices, makeTestIndexWithStore(name, 50-i, "2000"))
					}
					return indices
				}(),
				nodes: CatNode{
					{Name: "data-1", NodeRole: "h", DiskTotal: "1gb", DiskUsedPercent: "85", DiskUsed: "0.85gb", DiskAvailable: "0.15gb"},
				},
			},
			run: func() []string {
				return Filter_of_filter([]string{"pattern", "water_level"}, []structs.Filter{
					{Filtertype: "pattern", Kind: "prefix", Value: []string{"logstash-"}},
					{Filtertype: "water_level", UpperLimit: 80, LowerLimit: 79},
				})
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			restore := setMetadataProviderForTest(tc.provider)
			defer restore()

			actual := sorted(tc.run())
			expectedPath := filepath.Join("testdata", tc.file)

			raw, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("failed to read golden file %s: %v", expectedPath, err)
			}

			var expected []string
			if err := json.Unmarshal(raw, &expected); err != nil {
				t.Fatalf("failed to unmarshal golden file %s: %v", expectedPath, err)
			}

			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("golden mismatch for %s: got %v want %v", tc.name, actual, expected)
			}
		})
	}
}
