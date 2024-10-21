// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package filewalk

import "testing"
func TestXxx(t *testing.T) {
	walk :=NewFilewalk("/chrome/buildcef/chromium/src")		
	walk.Walk()
}