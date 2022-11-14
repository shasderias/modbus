package modbus

import (
	"reflect"
	"testing"
)

func TestBytesToBools(t *testing.T) {
	tests := []struct {
		name  string
		slice []byte
		want  []bool
	}{
		{"EmptySlice", []byte{}, []bool{}},
		{
			"OneByte",
			[]byte{0b00000001},
			[]bool{true, false, false, false, false, false, false, false},
		},
		{"TwoBytes",
			[]byte{0b00100000, 0b00000001},
			[]bool{
				false, false, false, false, false, true, false, false,
				true, false, false, false, false, false, false, false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bytesToBools(tt.slice); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("bytesToBools() = %v, want %v", got, tt.want)
			}
		})
	}
}
