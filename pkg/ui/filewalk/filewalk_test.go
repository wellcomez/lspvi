package filewalk

import "testing"
func TestXxx(t *testing.T) {
	walk :=NewFilewalk("/chrome/buildcef/chromium/src")		
	walk.Walk()
}