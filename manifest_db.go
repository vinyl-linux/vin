package main

import (
	_ "log"

	"github.com/hashicorp/go-memdb"
	"github.com/hashicorp/go-version"
)

var (
	schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"package": &memdb.TableSchema{
				Name: "package",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.CompoundIndex{Indexes: []memdb.Indexer{&memdb.StringFieldIndex{Field: "Provides"}, &memdb.StringFieldIndex{Field: "Version"}}},
					},
					"provides": &memdb.IndexSchema{
						Name:    "provides",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Provides"},
					},
					"version": &memdb.IndexSchema{
						Name:    "version",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Version"},
					},
				},
			},
		},
	}
)

type ManifestDB struct {
	db *memdb.MemDB
}

func LoadDB() (d ManifestDB, err error) {
	// Get manifests
	manifests, err := Manifests()
	if err != nil {
		return
	}

	d.db, err = memdb.NewMemDB(schema)
	if err != nil {
		return
	}

	tx := d.db.Txn(true)
	for _, manifest := range manifests {
		err = tx.Insert("package", &manifest)
		if err != nil {
			return
		}
	}

	tx.Commit()

	return
}

// Satisfies returns a slice of Manifests which satisifes a constraint
func (d ManifestDB) Satisfies(pkg string, constraint version.Constraints) (s []*Manifest, err error) {
	s = make([]*Manifest, 0)

	tx := d.db.Txn(false)
	defer tx.Abort()

	it, err := tx.Get("package", "provides", pkg)
	if err != nil {
		return
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		m := obj.(*Manifest)

		if constraint.Check(m.Version) {
			s = append(s, m)
		}
	}

	return
}
