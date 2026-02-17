package tui

import (
	"testing"
)

func TestDefaultGlobalKeyMap(t *testing.T) {
	km := DefaultGlobalKeyMap()

	if len(km.Quit.Keys()) == 0 {
		t.Error("Quit binding has no keys")
	}
	if len(km.Back.Keys()) == 0 {
		t.Error("Back binding has no keys")
	}
	if len(km.Help.Keys()) == 0 {
		t.Error("Help binding has no keys")
	}
}
