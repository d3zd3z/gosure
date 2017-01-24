package sure

// Migrate hashes from oldTree to newTree.  Any files that are the
// same in the oldTree as the newTree will have their hash migrated to
// the new tree.
func MigrateHashes(oldTree, newTree *Tree) {
	// TODO: Although doing nothing is valid, it means update is
	// always going to be slightly slower.
}
