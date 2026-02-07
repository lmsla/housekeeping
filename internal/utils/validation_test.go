package utils

import (
	"housekeeping/internal/structs"
	"testing"
)

// TestValidateFilters_P1_DeleteIndicesWithoutPattern 測試 delete_indices 必須有 pattern filter
func TestValidateFilters_P1_DeleteIndicesWithoutPattern(t *testing.T) {
	action := structs.Actiond{
		Action: "delete_indices",
		Filters: []structs.Filter{
			{
				Filtertype: "age",
				Direction:  "older",
				Unit:       "days",
				UnitCount:  30,
			},
		},
	}

	err := validateFilters(0, action)
	if err == nil {
		t.Error("Expected error for delete_indices without pattern filter, but got nil")
	}
	if err != nil && err.Error() != "action[0]: delete_indices action 必須配置 pattern filter 以避免意外刪除系統索引" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestValidateFilters_P1_DeleteIndicesWithPattern 測試 delete_indices 配合 pattern filter 應該通過
func TestValidateFilters_P1_DeleteIndicesWithPattern(t *testing.T) {
	action := structs.Actiond{
		Action: "delete_indices",
		Filters: []structs.Filter{
			{
				Filtertype: "pattern",
				Kind:       "prefix",
				Value:      []string{"logs-"},
			},
			{
				Filtertype: "age",
				Direction:  "older",
				Unit:       "days",
				UnitCount:  30,
			},
		},
	}

	err := validateFilters(0, action)
	if err != nil {
		t.Errorf("Expected no error for delete_indices with pattern filter, but got: %v", err)
	}
}

// TestValidateFilters_P1_SpaceWithoutPattern 測試 space filter 必須有 pattern
func TestValidateFilters_P1_SpaceWithoutPattern(t *testing.T) {
	action := structs.Actiond{
		Action: "close", // 使用 close action 來測試 space filter
		Filters: []structs.Filter{
			{
				Filtertype: "space",
				DiskSpace:  100,
			},
		},
	}

	err := validateFilters(0, action)
	if err == nil {
		t.Error("Expected error for space filter without pattern, but got nil")
	}
	if err != nil && err.Error() != "action[0]: space filter 必須配合 pattern filter 使用，否則會影響所有索引（包括系統索引）" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestValidateFilters_P1_WaterLevelWithoutPattern 測試 water_level filter 必須有 pattern
func TestValidateFilters_P1_WaterLevelWithoutPattern(t *testing.T) {
	action := structs.Actiond{
		Action: "close", // 使用 close action 來測試 water_level filter
		Filters: []structs.Filter{
			{
				Filtertype: "water_level",
				UpperLimit: 80,
			},
		},
	}

	err := validateFilters(0, action)
	if err == nil {
		t.Error("Expected error for water_level filter without pattern, but got nil")
	}
	if err != nil && err.Error() != "action[0]: water_level filter 必須配合 pattern filter 使用，否則會影響所有索引（包括系統索引）" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestValidateFilters_P1_NodeRoleWithoutPattern 測試 node_role filter 必須有 pattern
func TestValidateFilters_P1_NodeRoleWithoutPattern(t *testing.T) {
	action := structs.Actiond{
		Action: "close",
		Filters: []structs.Filter{
			{
				Filtertype: "node_role",
				Value:      []string{"w"},
			},
		},
	}

	err := validateFilters(0, action)
	if err == nil {
		t.Error("Expected error for node_role filter without pattern, but got nil")
	}
	if err != nil && err.Error() != "action[0]: node_role filter 必須配合 pattern filter 使用，以限制操作範圍" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestValidateFilters_P1_NodeRoleWithPattern 測試 node_role + pattern 應該通過
func TestValidateFilters_P1_NodeRoleWithPattern(t *testing.T) {
	action := structs.Actiond{
		Action: "close",
		Filters: []structs.Filter{
			{
				Filtertype: "node_role",
				Value:      []string{"w"},
			},
			{
				Filtertype: "pattern",
				Kind:       "prefix",
				Value:      []string{"logs-"},
			},
		},
	}

	err := validateFilters(0, action)
	if err != nil {
		t.Errorf("Expected no error for node_role with pattern filter, but got: %v", err)
	}
}

// TestValidateFilters_P1_SpaceWithPattern 測試 space + pattern 應該通過
func TestValidateFilters_P1_SpaceWithPattern(t *testing.T) {
	action := structs.Actiond{
		Action: "delete_indices",
		Filters: []structs.Filter{
			{
				Filtertype: "pattern",
				Kind:       "prefix",
				Value:      []string{"logs-"},
			},
			{
				Filtertype: "space",
				DiskSpace:  100,
			},
		},
	}

	err := validateFilters(0, action)
	if err != nil {
		t.Errorf("Expected no error for space with pattern filter, but got: %v", err)
	}
}

// TestValidateFilters_P0_AgeAndSpace 測試 age + space 互斥（P0 已修復）
func TestValidateFilters_P0_AgeAndSpace(t *testing.T) {
	action := structs.Actiond{
		Action: "delete_indices",
		Filters: []structs.Filter{
			{
				Filtertype: "pattern",
				Kind:       "prefix",
				Value:      []string{"logs-"},
			},
			{
				Filtertype: "age",
				Direction:  "older",
				Unit:       "days",
				UnitCount:  30,
			},
			{
				Filtertype: "space",
				DiskSpace:  100,
			},
		},
	}

	err := validateFilters(0, action)
	if err == nil {
		t.Error("Expected error for age + space combination, but got nil")
	}
	if err != nil && err.Error() != "action[0]: age 和 space 過濾器不能同時使用" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestValidateFilters_P0_AgeAndWaterLevel 測試 age + water_level 互斥（P0 已修復）
func TestValidateFilters_P0_AgeAndWaterLevel(t *testing.T) {
	action := structs.Actiond{
		Action: "delete_indices",
		Filters: []structs.Filter{
			{
				Filtertype: "pattern",
				Kind:       "prefix",
				Value:      []string{"logs-"},
			},
			{
				Filtertype: "age",
				Direction:  "older",
				Unit:       "days",
				UnitCount:  30,
			},
			{
				Filtertype: "water_level",
				UpperLimit: 80,
			},
		},
	}

	err := validateFilters(0, action)
	if err == nil {
		t.Error("Expected error for age + water_level combination, but got nil")
	}
	if err != nil && err.Error() != "action[0]: age 和 water_level 過濾器不能同時使用" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestValidateFilters_P0_SpaceAndWaterLevel 測試 space + water_level 互斥（P0 已修復）
func TestValidateFilters_P0_SpaceAndWaterLevel(t *testing.T) {
	action := structs.Actiond{
		Action: "delete_indices",
		Filters: []structs.Filter{
			{
				Filtertype: "pattern",
				Kind:       "prefix",
				Value:      []string{"logs-"},
			},
			{
				Filtertype: "space",
				DiskSpace:  100,
			},
			{
				Filtertype: "water_level",
				UpperLimit: 80,
			},
		},
	}

	err := validateFilters(0, action)
	if err == nil {
		t.Error("Expected error for space + water_level combination, but got nil")
	}
	if err != nil && err.Error() != "action[0]: space 和 water_level 過濾器不能同時使用" {
		t.Errorf("Unexpected error message: %v", err)
	}
}
