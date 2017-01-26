Gosure file integrity
*********************

It has been said that backups aren't useful unless you've tested them.
But, how does one know that a test restore actually worked?  Gosure is
designed to help with this.

History
=======

The md5sum program captures the MD5 hash of a set of files.  It can
also read this output and compare the hashes against the files.  By
capturing the hashes before the backup, and comparing them after a
test restore, you can gain a bit of confidence that the contents of
files is at least correct.

However, this doesn't capture the permissions and other attributes of
the files.  Sometimes a restore can fail for this kind of reason.

Intrusion detection
===================

There have been several similar solutions focused on intrusion
detection.  Tripwire and FreeVeracity (or Veracity) come to mind.  The
idea is that the files are compared in place to verify that nobody has
modified them.

Unfortunately, at least tripwire seems to focus so heavily on this
intrusion detection problem, that the tool doesn't work very well for
verifying backups.  It really wants a central database, and to use
files by absolute pathname.  FreeVeracity was quite useful for
verifying backups, however, it appears to have vanished entirely (it
was under an unusual license).

Incremental updates
-------------------

One thing that none of these solutions addressed was that of
incremental updates, probably because of the focus on intrusion
detection.  In a normal running system, the POSIX *ctime* field can be
reliably used to determine if a file has been modified.  By making use
of this, the integrity program can avoid recomputing hashes of files
that haven't changed.  This strategy is similar to what most backup
software does as well.  This is important, because taking the time to
hash every file can make the integrity update take so long that people
avoid running it.  Full hashing is impractical for the same reasons
that regular full backups are usually impractical.

Using gosure
************

Getting it
==========

Gosure is written in Go_.

.. _Go: https://golang.org/

It is generally best to work with Go using its idea of a workspace.
You should create a directory somewhere for go work, and set the
environment variable ``GOPATH`` to point to this.  Once this is done,
use the go tools to fetch this project::

    $ go get davidb.org/code/gosure

.. note::
   Although this project is hosted at github.com (currently), the go
   tool should complain if you try to fetch using that path.  This is
   because the package needs to be able to reference sub-packages by
   full name, and these will only work if the package is fetched via
   its canonical name.

Once the tree is present::

    $ go install davidb.org/code/gosure

should install the gosure program itself in ``$GOPATH/bin``.  Add this
to the path to make things more convenient.  The execuable is
standalone, and has no dependencies on the source tree.

Basic usage
===========

Change to a directory you wish to keep integrity for, for example, my
home directory::

    $ cd
    $ gosure scan

This will scan the filesystem (possibly showing progress), and leave a
``2sure.dat.gz`` (the 2sure is historical, FreeVeracity used a name
starting with a 0, and having the digit makes it near the beginning of
a directory listing).  You can view this file if you'd like.  Aside
from being compressed, the format is plain ASCII (even if your
filenames are not).

Then you can do::

    $ gosure check

to verify the directory.  This will show any differences.  If you back
up this file with your data, you can run ``gosure`` after a restore to
check if the backup is correct.

Later, you can run ::

    $ gosure update

which will move the ``2sure.dat.gz`` file to ``2sure.bak.gz``, and refresh
the hashes of any files that have changed.  After you have these two
files::

    $ gosure signoff

will compare the old scan with the current, and report on what has
changed between them.
