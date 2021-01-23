package main

import (
	_ "log"

	"fmt"
	"sort"

	"github.com/hashicorp/go-memdb"
	"github.com/hashicorp/go-version"
)

var (
	schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"package": {
				Name: "package",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"provides": {
						Name:    "provides",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Provides"},
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
		err = tx.Insert("package", manifest)
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

// Latest version returns the latest version of a Manifest to saitisfy a constraint
func (d ManifestDB) Latest(pkg string, constraint version.Constraints) (m *Manifest, err error) {
	//log.Printf("pkg: %+v, constraint: %+v", pkg, constraint)

	satisfiers, err := d.Satisfies(pkg, constraint)
	if err != nil {
		return
	}

	if len(satisfiers) == 0 {
		err = fmt.Errorf("nothing satisfies %s %s", pkg, constraint.String())

		return
	}

	sort.Sort(ManifestByVersion(satisfiers))

	m = satisfiers[len(satisfiers)-1]

	return
}
