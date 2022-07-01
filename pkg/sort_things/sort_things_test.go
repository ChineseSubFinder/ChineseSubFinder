package sort_things

import (
	"reflect"
	"testing"
)

func TestSortStringSliceByLength(t *testing.T) {
	type args struct {
		m []string
	}
	tests := []struct {
		name string
		args args
		want PathSlices
	}{
		{name: "00", args: args{m: []string{"1", "12", "123"}}, want: PathSlices{
			PathSlice{"123"},
			PathSlice{"12"},
			PathSlice{"1"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SortStringSliceByLength(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SortStringSliceByLength() = %v, want %v", got, tt.want)
			}
		})
	}
}
