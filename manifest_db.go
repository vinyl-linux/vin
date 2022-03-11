package main

import (
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
	d.db, err = memdb.NewMemDB(schema)
	if err != nil {
		return
	}

	err = d.loadManifests()

	return
}

func (d *ManifestDB) Reload() error {
	return d.loadManifests()
}

// Satisfies returns a slice of Manifests which satisfies a constraint
func (d ManifestDB) Satisfies(pkg string, constraint version.Constraints) (s []*Manifest, err error) {
	s = make([]*Manifest, 0)

	tx := d.db.Txn(false)
	defer tx.Abort()

	it, err := tx.Get("package", "provides", pkg)
	if err != nil {
		return
	}

	var (
		latestManifest *Manifest
	)

	for obj := it.Next(); obj != nil; obj = it.Next() {
		m := obj.(*Manifest)

		if constraint.String() == latest.String() && (latestManifest == nil || m.Version.GreaterThan(latestManifest.Version)) {
			latestManifest = m
		}

		if constraint.Check(m.Version) {
			s = append(s, m)
		}
	}

	if len(s) == 0 && latestManifest != nil {
		s = []*Manifest{latestManifest}
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

func (d *ManifestDB) loadManifests() (err error) {
	// Get manifests
	manifests, err := Manifests()
	if err != nil {
		return
	}

	// empty memdb
	tx := d.db.Txn(true)

	_, err = tx.DeleteAll("package", "id")
	if err != nil {
		return
	}

	for _, manifest := range manifests {
		err = d.addManifest(tx, manifest)
		if err != nil {
			return
		}
	}

	tx.Commit()

	return
}

func (d *ManifestDB) addManifest(tx *memdb.Txn, manifest *Manifest) error {
	return tx.Insert("package", manifest)
}

func (d *ManifestDB) deleteManifest(name string) (err error) {
	m, err := d.Latest(name, latest)
	if err != nil {
		return
	}

	tx := d.db.Txn(true)
	defer tx.Commit()

	return tx.Delete("package", m)
}
