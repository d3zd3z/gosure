# Build my stuff.

include $(GOROOT)/src/Make.inc

TARG = gosure
GOFILES = $(wildcard *.go)

include $(GOROOT)/src/Make.cmd
