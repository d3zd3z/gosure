# Build my stuff.

include $(GOROOT)/src/Make.inc

TARG = gosure
GOFILES := gosure.go compare.go localtree.go surefile.go

# The OpenSSL sha1 library is about 4.5 times faster on my machine,
# but using it is more complex, and has complicated licensing
# requirements.  Another option would be the git sha1 code, which is
# optimized for gcc.
# USE_OPENSSL = false

ifdef USE_OPENSSL
GOFILES += hashfile-openssl.go
else
GOFILES += hashfile-go.go
endif

include $(GOROOT)/src/Make.cmd
