package main

import (
	"fmt"

	"github.com/hashicorp/go-version"
)

var (
	latest, _ = version.NewConstraint(">= 0")
)

type Graph struct {
	mdb            *ManifestDB
	sdb            *StateDB
	output         chan string
	depConstraints map[string]version.Constraints
	resolved       map[string]*Manifest
	seen           map[string]*Manifest
	tasks          []*Manifest
	level          int
}

func NewGraph(mdb *ManifestDB, sdb *StateDB, output chan string) *Graph {
	return &Graph{
		mdb:            mdb,
		sdb:            sdb,
		output:         output,
		depConstraints: make(map[string]version.Constraints),
		resolved:       make(map[string]*Manifest),
		seen:           make(map[string]*Manifest),
		tasks:          make([]*Manifest, 0),
		level:          0,
	}
}

// restartReset will remove all found manifests when restarting
// this ensures that re-forming the graph with smarter constraints wont
// throw lots of daft errors
func (g *Graph) restartReset() {
	g.output <- "(re)starting dependency graph solution"

	g.resolved = make(map[string]*Manifest)
	g.seen = make(map[string]*Manifest)
	g.tasks = make([]*Manifest, 0)
}

func (g *Graph) getManifest(pkg string, con version.Constraints) (m *Manifest, err error) {
	// Ensure depConstraints has a valid set of constraints for this pkg?
	_, ok := g.depConstraints[pkg]
	if !ok {
		g.depConstraints[pkg] = con
	} else {
		g.depConstraints[pkg] = append(g.depConstraints[pkg], con...)
	}

	// Get latest version that supports this updated constraint
	return g.mdb.Latest(pkg, g.depConstraints[pkg])
}

// Solve takes a package, version constraint (or nil, which is 'latest') and returns
// a set of manifests to install, in order.
func (g *Graph) Solve(profile, pkg string, con version.Constraints) (tasks []*Manifest, err error) {
	var m *Manifest

	if con == nil {
		m, err = g.mdb.Latest(pkg, latest)
	} else {
		m, err = g.mdb.Latest(pkg, con)
	}

	if err != nil {
		return
	}

	g.restartReset()
	g.depConstraints = make(map[string]version.Constraints)

	_, err = g.solve(m, profile)
	tasks = g.tasks

	return
}

func (g *Graph) solve(m *Manifest, profile string) (restart bool, err error) {
	g.level++
	defer func() {
		if restart {
			g.restartReset()
		}

		g.level--
	}()

	g.output <- fmt.Sprintf("resolving %q (%q)\n", m.Provides, m.VersionStr)

	g.seen[m.Provides] = m

	prf, ok := m.Profiles[profile]
	if !ok {
		err = fmt.Errorf("%s: unknown profile %q", m.Provides, profile)
	}

	var m1 *Manifest
	for i := 0; i < len(prf.Deps); i++ {
		// We use this for-loop prototype, rather than `for _, dep := range prf.Deps` so
		// that we can easily restart the loop when we need to re-calculate dependencies
		dep := prf.Deps[i]

		pkg := dep.Package()
		con, _ := dep.Constraint() // we can ignore the constraint error- if we get this far, it's tested

		m1, err = g.getManifest(pkg, con)
		if err != nil {
			return
		}

		// check whether we've already processed this dependency. If we have,
		// and the version we've stored is different to the one we just picked up
		// we need to remove it from 'seen' and restart.
		//
		// This means next time we iterate through our graph, we'll get the right
		// version first time (or... earlier, depending how often this dep is in the
		// graph and is wrong)
		prv, seen := g.seen[pkg]
		if seen {
			if prv.String() != m1.String() {
				delete(g.seen, pkg)

				if g.level > 1 {
					restart = true

					return
				}

				// we're at the top level, so let's restart!
				i = -1

				// delete m1 from tasks, resolved, seen
				g.restartReset()
				continue
			}

			_, resolved := g.resolved[pkg]
			if !resolved {
				err = fmt.Errorf("circular dependency: %q -> %q", m.Provides, pkg)

				return
			}
		}

		restart, err = g.solve(m1, profile)
		if err != nil {
			return
		}

		if restart {
			if g.level > 1 {
				return
			}

			// Start the wider for loop at the beginning
			i = 0

			continue
		}

	}

	if !in(g.tasks, m) {
		g.tasks = append(g.tasks, m)
	}

	g.resolved[m.Provides] = m

	return
}

func in(s []*Manifest, m *Manifest) bool {
	for _, m1 := range s {
		if m1 == m {
			return true
		}
	}

	return false
}
