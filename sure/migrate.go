package sure

// Notice: there is a fairly strong assumption that this tool will not
// be used to cross filesystems.  We don't store device numbers and
// can't distinguish inodes from different filesystems.

type inoMap map[uint64]*RegAtts

// Migrate hashes from oldTree to newTree.  Any files that are the
// same in the oldTree as the newTree will have their hash migrated to
// the new tree.
func MigrateHashes(oldTree, newTree *Tree) {
	oldHashes := make(inoMap)
	getHashes(oldTree, oldHashes)
	// log.Printf("%d hashes in old tree", len(oldHashes))
	updateHashes(newTree, oldHashes)
}

// Walk through the tree, gathering all of the hashes of existing
// nodes, keeping a pointer to the attribute map of the node.  We only
// use the inode to distinguish the nodes, so the position and such of
// the node is not needed.
func getHashes(tree *Tree, hashes inoMap) {
	// Walk the children
	for _, c := range tree.Children {
		getHashes(c, hashes)
	}

	// Then the file nodes.
	for _, f := range tree.Files {
		atts, ok := f.Atts.(*RegAtts)

		// Only attend to ones with a 'sha1' property.
		if !ok || atts.Sha1 == nil {
			continue
		}

		// TODO: If the node is already present in the map,
		// ideally the attributes should be the same.  If not,
		// it suggests the file changed while being loaded.
		// Come up with a way of dealing with this or just
		// warning.

		hashes[atts.Ino] = atts
	}
}

// Update any hashes that have the same attributes.
func updateHashes(tree *Tree, hashes inoMap) {
	// Walk the children
	for _, c := range tree.Children {
		updateHashes(c, hashes)
	}

	// Then the file nodes.
	for _, f := range tree.Files {
		atts, ok := f.Atts.(*RegAtts)

		if !ok {
			continue
		}

		oldAtt, ok := hashes[atts.Ino]
		if !ok {
			continue
		}

		// For sanity, make sure the ctime, and size are the
		// same.
		if oldAtt.Ctime != atts.Ctime || oldAtt.Size != atts.Size {
			continue
		}

		atts.Sha1 = oldAtt.Sha1
	}
}
