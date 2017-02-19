package kenken

import (
	"reflect"
	"testing"
)

func TestDecompressSGTString(t *testing.T) {
	cases := []struct {
		compressed   string
		uncompressed string
	}{
		{"", ""},
		{"abc", "abc"},
		{"abc,foo", "abc,foo"},
		{"a3,foo", "aaa,foo"},
		{"a_a3_a_", "a_aaa_a_"},
		{"a25", "aaaaaaaaaaaaaaaaaaaaaaaaa"},
		{"a__a_a_7a__aa_5a_5aa_4a_5a__a_a3_a__b_5a_a_3aababb_a6", "a__a_a_______a__aa_____a_____aa____a_____a__a_aaa_a__b_____a_a___aababb_aaaaaa"},
	}
	for _, c := range cases {
		d := decompressSGTString(c.compressed)
		if d != c.uncompressed {
			t.Errorf("decompressSGTString(%q) == %q, want %q", c.compressed, d, c.uncompressed)
		}
	}
}

func TestReadSGT(t *testing.T) {
	cases := []struct {
		id     string
		soln   string
		puzzle Puzzle
	}{
		{"3:a_a3_a_,a3m6m3s1",
			"2 1 3\n3 2 1\n1 3 2\n",
			Puzzle{
				Answer: [][]int{
					[]int{2, 1, 3},
					[]int{3, 2, 1},
					[]int{1, 3, 2},
				},
				Clue: [][]int{
					[]int{3, 0, 6},
					[]int{3, 0, 0},
					[]int{0, 1, 0},
				},
				Operation: [][]Operation{
					[]Operation{2, 0, 4},
					[]Operation{4, 0, 0},
					[]Operation{0, 3, 0},
				},
				Vertical: [][]bool{
					[]bool{false, true},
					[]bool{true, false},
					[]bool{true, false},
				},
				Horizontal: [][]bool{
					[]bool{true, false},
					[]bool{true, true},
					[]bool{false, true},
				},
			},
		},
		{"6:aa_a_6ba_aa_3a_4a3_3a__aab__b__a,d3m24d3s3m12a8a6a17m6s2a6s3s1d2m6d2",
			"6 2 4 1 3 5\n3 4 5 6 1 2\n2 1 3 5 6 4\n4 6 1 2 5 3\n1 5 2 3 4 6\n5 3 6 4 2 1\n",
			Puzzle{
				Answer: [][]int{
					[]int{6, 2, 4, 1, 3, 5},
					[]int{3, 4, 5, 6, 1, 2},
					[]int{2, 1, 3, 5, 6, 4},
					[]int{4, 6, 1, 2, 5, 3},
					[]int{1, 5, 2, 3, 4, 6},
					[]int{5, 3, 6, 4, 2, 1},
				},
				Clue: [][]int{
					[]int{3, 0, 24, 0, 3, 3},
					[]int{12, 0, 8, 0, 0, 0},
					[]int{6, 0, 0, 17, 0, 0},
					[]int{0, 6, 0, 0, 2, 0},
					[]int{6, 3, 0, 1, 2, 6},
					[]int{0, 2, 0, 0, 0, 0},
				},
				Operation: [][]Operation{
					[]Operation{5, 0, 4, 0, 5, 3},
					[]Operation{4, 0, 2, 0, 0, 0},
					[]Operation{2, 0, 0, 2, 0, 0},
					[]Operation{0, 4, 0, 0, 3, 0},
					[]Operation{2, 3, 0, 3, 5, 4},
					[]Operation{0, 5, 0, 0, 0, 0},
				},
				Vertical: [][]bool{
					[]bool{false, true, false, true, true},
					[]bool{false, true, true, true, true},
					[]bool{true, true, true, false, false},
					[]bool{true, false, true, true, false},
					[]bool{true, false, true, true, true},
					[]bool{true, false, true, true, true},
				},
				Horizontal: [][]bool{
					[]bool{true, true, false, true, false},
					[]bool{true, false, true, true, true},
					[]bool{true, false, true, true, true},
					[]bool{false, true, false, true, false},
					[]bool{false, true, true, true, false},
					[]bool{false, true, true, true, false},
				},
			},
		},
		{"8:a__a_a_7a__aa_5a_5aa_4a_5a__a_a3_a__b_5a_a_3aababb_a6,s5s2d2d4s3a7m56a11a7d3m16s2a14s2m20a18a10a9m7m14s3m252d3d3s2s1m10m320d4",
			"8 3 7 2 1 4 5 6\n7 8 5 4 3 6 2 1\n6 2 3 5 8 1 7 4\n2 1 8 6 5 7 4 3\n5 7 1 8 4 3 6 2\n4 5 2 3 6 8 1 7\n1 4 6 7 2 5 3 8\n3 6 4 1 7 2 8 5\n",
			Puzzle{
				Answer: [][]int{
					[]int{8, 3, 7, 2, 1, 4, 5, 6},
					[]int{7, 8, 5, 4, 3, 6, 2, 1},
					[]int{6, 2, 3, 5, 8, 1, 7, 4},
					[]int{2, 1, 8, 6, 5, 7, 4, 3},
					[]int{5, 7, 1, 8, 4, 3, 6, 2},
					[]int{4, 5, 2, 3, 6, 8, 1, 7},
					[]int{1, 4, 6, 7, 2, 5, 3, 8},
					[]int{3, 6, 4, 1, 7, 2, 8, 5},
				},
				Clue: [][]int{
					[]int{5, 0, 2, 2, 4, 0, 3, 7},
					[]int{56, 0, 0, 0, 11, 7, 0, 0},
					[]int{3, 16, 2, 0, 0, 0, 14, 0},
					[]int{0, 0, 0, 2, 20, 18, 10, 0},
					[]int{9, 7, 0, 0, 0, 0, 0, 14},
					[]int{0, 3, 0, 252, 0, 0, 3, 0},
					[]int{3, 2, 1, 0, 0, 10, 0, 320},
					[]int{0, 0, 4, 0, 0, 0, 0, 0},
				},
				Operation: [][]Operation{
					[]Operation{3, 0, 3, 5, 5, 0, 3, 2},
					[]Operation{4, 0, 0, 0, 2, 2, 0, 0},
					[]Operation{5, 4, 3, 0, 0, 0, 2, 0},
					[]Operation{0, 0, 0, 3, 4, 2, 2, 0},
					[]Operation{2, 4, 0, 0, 0, 0, 0, 4},
					[]Operation{0, 3, 0, 4, 0, 0, 5, 0},
					[]Operation{5, 3, 3, 0, 0, 4, 0, 4},
					[]Operation{0, 0, 5, 0, 0, 0, 0, 0},
				},
				Vertical: [][]bool{
					[]bool{false, true, true, true, false, true, true},
					[]bool{false, true, true, true, true, true, true},
					[]bool{true, true, false, true, true, true, false},
					[]bool{true, false, true, true, true, true, true},
					[]bool{true, false, true, true, true, true, true},
					[]bool{true, false, true, false, true, true, true},
					[]bool{true, true, false, true, true, true, true},
					[]bool{true, true, false, true, true, true, false},
				},
				Horizontal: [][]bool{
					[]bool{true, true, false, true, false, true, false},
					[]bool{true, true, false, true, true, true, false},
					[]bool{false, true, true, true, true, true, true},
					[]bool{false, true, true, false, true, true, true},
					[]bool{true, false, true, false, true, false, false},
					[]bool{true, false, true, false, false, true, false},
					[]bool{false, true, true, false, true, false, true},
					[]bool{false, true, false, true, false, true, false},
				},
			},
		},
	}
	for _, c := range cases {
		p, err := ReadSGT(c.id, c.soln)
		if err != nil {
			t.Errorf("ReadSGT(%q, %q) error: %v, want %+v", c.id, c.soln, err, c.puzzle)
			continue
		}
		if !reflect.DeepEqual(*p, c.puzzle) {
			t.Errorf("ReadSGT(%q, %q) == %+v, want %+v", c.id, c.soln, *p, c.puzzle)
		}
	}
}
