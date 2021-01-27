package main

import (
	"reflect"
	"testing"
)

func TestStateDB_Manifest(t *testing.T) {
	stateDB = "testdata/vin.db"

	s, err := LoadStateDB()
	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}

	expect := Manifest{
		ID:         "world 1257894000",
		Provides:   "world",
		VersionStr: "1257894000",
		Profiles:   map[string]Profile{"default": Profile{Deps: []Dep{}}},
	}

	got, err := s.Meta()
	got.Version = nil

	if !reflect.DeepEqual(expect, got) {
		t.Errorf("expected %#v, recveived %#v", expect, got)
	}
}
