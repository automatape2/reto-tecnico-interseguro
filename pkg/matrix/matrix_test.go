package matrix

import "testing"

func TestRowsCols(t *testing.T) {
	m := Matrix{{1, 2, 3}, {4, 5, 6}}
	if got := m.Rows(); got != 2 {
		t.Errorf("Rows() = %d, want 2", got)
	}
	if got := m.Cols(); got != 3 {
		t.Errorf("Cols() = %d, want 3", got)
	}
}

func TestRowsColsEmpty(t *testing.T) {
	var m Matrix
	if got := m.Rows(); got != 0 {
		t.Errorf("Rows() on nil matrix = %d, want 0", got)
	}
	if got := m.Cols(); got != 0 {
		t.Errorf("Cols() on nil matrix = %d, want 0", got)
	}
}

func TestIsRectangular(t *testing.T) {
	cases := []struct {
		name string
		m    Matrix
		want bool
	}{
		{"square", Matrix{{1, 2}, {3, 4}}, true},
		{"tall", Matrix{{1}, {2}, {3}}, true},
		{"wide", Matrix{{1, 2, 3}}, true},
		{"ragged", Matrix{{1, 2}, {3}}, false},
		{"empty rows", Matrix{}, false},
		{"empty cols", Matrix{{}}, false},
		{"nil", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.m.IsRectangular(); got != tc.want {
				t.Errorf("IsRectangular() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	if err := (Matrix{{1, 2}, {3, 4}}).Validate(); err != nil {
		t.Errorf("Validate() on well-formed matrix returned %v, want nil", err)
	}
	if err := (Matrix{}).Validate(); err != ErrEmptyMatrix {
		t.Errorf("Validate() on empty matrix = %v, want ErrEmptyMatrix", err)
	}
	if err := (Matrix{{}}).Validate(); err != ErrEmptyMatrix {
		t.Errorf("Validate() on zero-width matrix = %v, want ErrEmptyMatrix", err)
	}
	if err := (Matrix{{1, 2}, {3}}).Validate(); err != ErrRaggedMatrix {
		t.Errorf("Validate() on ragged matrix = %v, want ErrRaggedMatrix", err)
	}
}

func TestCloneIndependence(t *testing.T) {
	original := Matrix{{1, 2}, {3, 4}}
	clone := original.Clone()

	clone[0][0] = 999

	if original[0][0] != 1 {
		t.Errorf("mutating clone affected original: got %v, want 1", original[0][0])
	}
	if clone[0][0] != 999 {
		t.Errorf("clone was not mutated as expected: got %v, want 999", clone[0][0])
	}
}

func TestIdentity(t *testing.T) {
	id := Identity(3)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			want := 0.0
			if i == j {
				want = 1.0
			}
			if id[i][j] != want {
				t.Errorf("Identity(3)[%d][%d] = %v, want %v", i, j, id[i][j], want)
			}
		}
	}
}
