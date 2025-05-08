package filter

import (
	"github.com/fxnoob/bunql/dto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseFilterParam(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		op            string
		value         interface{}
		logic         string
		expectedGroup dto.FilterGroup
		expectError   bool
	}{
		{
			name:  "Valid filter with string value and default logic",
			key:   "name",
			op:    "eq",
			value: "John",
			logic: "",
			expectedGroup: dto.FilterGroup{
				Logic: "and",
				Filters: []dto.Filter{
					{
						Field:    "name",
						Operator: "eq",
						Value:    "John",
					},
				},
				Groups: []dto.FilterGroup{},
			},
			expectError: false,
		},
		{
			name:  "Valid filter with integer value and OR logic",
			key:   "age",
			op:    "gt",
			value: 30,
			logic: "or",
			expectedGroup: dto.FilterGroup{
				Logic: "or",
				Filters: []dto.Filter{
					{
						Field:    "age",
						Operator: "gt",
						Value:    30,
					},
				},
				Groups: []dto.FilterGroup{},
			},
			expectError: false,
		},
		{
			name:  "Valid filter with array value",
			key:   "status",
			op:    "in",
			value: []string{"active", "pending"},
			logic: "and",
			expectedGroup: dto.FilterGroup{
				Logic: "and",
				Filters: []dto.Filter{
					{
						Field:    "status",
						Operator: "in",
						Value:    []string{"active", "pending"},
					},
				},
				Groups: []dto.FilterGroup{},
			},
			expectError: false,
		},
		{
			name:        "Invalid empty key",
			key:         "",
			op:          "eq",
			value:       "test",
			logic:       "and",
			expectError: true,
		},
		{
			name:        "Invalid operator",
			key:         "name",
			op:          "invalid",
			value:       "test",
			logic:       "and",
			expectError: true,
		},
		{
			name:  "Invalid logic defaults to AND",
			key:   "name",
			op:    "eq",
			value: "test",
			logic: "invalid",
			expectedGroup: dto.FilterGroup{
				Logic: "and",
				Filters: []dto.Filter{
					{
						Field:    "name",
						Operator: "eq",
						Value:    "test",
					},
				},
				Groups: []dto.FilterGroup{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := ParseFilterParam(tt.key, tt.op, tt.value, tt.logic)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedGroup.Logic, group.Logic)
				assert.Len(t, group.Filters, 1)
				assert.Equal(t, tt.expectedGroup.Filters[0].Field, group.Filters[0].Field)
				assert.Equal(t, tt.expectedGroup.Filters[0].Operator, group.Filters[0].Operator)
				assert.Equal(t, tt.expectedGroup.Filters[0].Value, group.Filters[0].Value)
				assert.Empty(t, group.Groups)
			}
		})
	}
}
