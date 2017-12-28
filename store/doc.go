// Package store manages the files that are used to hold sure data.
//
// Sure data is stored in a small number of files, referred to as sure
// files.  The features used will determine which of these files is
// used for a given operation.
//
// The most basic operation is to store the data in a pair of files.
// These will have names:
//
//    2sure.dat.gz
//    2sure.bak.gz
//
// with .dat.gz representing the most recent scan, and .bak.gz
// representing the previous one.  Earlier scans are discarded.
// During operation, the store will write to a temporary file
// 2sure.nn.gz, where nn is an incrementing integer, starting with 0.
// This is done because it is unsafe to overwrite files as root, and
// sure is often run as root.  The store will use the lowest numbered
// file that does not exist, and remove it when it is finished.  The
// use of the number prevents failure if a file is somehow left
// behind.
//
// The options available to the basic store are that the files may or
// may not be compressed (with or without the .gz suffix), the path
// that the files should be stored in (a general default is the base
// of the tree being scanned), and the prefix ("2sure") part of the
// name.
//
// When the backups are stored as weave files, the same file names are
// used (the software can tell by reading the files if they are weave
// format).  The primary difference being that the ".dat.gz" file will
// hold all of the previous versions.  The ".bak.gz" is redundant in
// this case, but one version is kept around for safety.  This format
// is generally more useful to accompany backups, for example, where
// the backup may have multiple versions stored in it, and these can
// be kept in the surefile store as well.
//
// In order to write the weave versions, the user must provide one or
// more tags to distinguish the particular version being written.
//
package store
