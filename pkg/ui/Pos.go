// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

type Pos struct {
	X, Y int
}

// LessThan returns true if b is smaller
func (l Pos) LessThan(b Pos) bool {
	if l.Y < b.Y {
		return true
	}
	if l.Y == b.Y && l.X < b.X {
		return true
	}
	return false
}

// GreaterThan returns true if b is bigger
func (l Pos) GreaterThan(b Pos) bool {
	if l.Y > b.Y {
		return true
	}
	if l.Y == b.Y && l.X > b.X {
		return true
	}
	return false
}

// GreaterEqual returns true if b is greater than or equal to b
func (l Pos) GreaterEqual(b Pos) bool {
	if l.Y > b.Y {
		return true
	}
	if l.Y == b.Y && l.X > b.X {
		return true
	}
	if l == b {
		return true
	}
	return false
}

// LessEqual returns true if b is less than or equal to b
func (l Pos) LessEqual(b Pos) bool {
	if l.Y < b.Y {
		return true
	}
	if l.Y == b.Y && l.X < b.X {
		return true
	}
	if l == b {
		return true
	}
	return false
}
