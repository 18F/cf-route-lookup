package main

import (
	"testing"
)

func Test_matchWildcard(t *testing.T) {
	type args struct {
		wildcard string
		route    string
	}
	tests := []struct {
		args    args
		want    bool
		wantErr bool
	}{
		{args{"ala.ma.kota", "ala.ma.kota"}, true, false},
		{args{"*.ma.kota", "ala.ma.kota"}, true, false},
		{args{"*.ma.*", "ala.ma.ko-ta"}, true, false},
		{args{"*.ma.*", "ala.ma.ko_ta"}, true, false},
		{args{"*.*.*", "ala.ma.kota"}, true, false},
		{args{"*.*", "ala.ma.kota"}, false, false},
		{args{"*.nie-ma.*", "ala.ma.kota"}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.args.wildcard+" vs "+tt.args.route, func(t *testing.T) {
			matched, err := matchWildcard(tt.args.wildcard, tt.args.route)
			if (err != nil) != tt.wantErr {
				t.Errorf("matchWildcard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if matched != tt.want {
				t.Errorf("matchWildcard() = %v, want %v", matched, tt.want)
			}
		})
	}
}
