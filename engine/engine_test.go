package engine

import "testing"

func TestAlignString(t *testing.T) {

	var giveStr, realStr, wantStr string

	giveStr = `123|4|5
a|d|ef`
	wantStr = `123|4|5
a  |d|ef`
	realStr = alignString("|", "|", giveStr)
	if realStr != wantStr {
		t.Errorf("align failed\ngive: %#v\nwant: %#v\nreal: %#v", giveStr, wantStr, realStr)
	}

	giveStr += "\n"
	wantStr += "\n"
	realStr = alignString("|", "|", giveStr)
	if realStr != wantStr {
		t.Errorf("align failed\ngive: %#v\nwant: %#v\nreal: %#v", giveStr, wantStr, realStr)
	}
}
