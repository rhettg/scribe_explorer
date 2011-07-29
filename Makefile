include $(GOROOT)/src/Make.inc

TARG=rangerweb
GOFILES=\
	tail.go\
	get_deep.go\
	rangerweb.go\
	filter.go\
	json_io.go

include $(GOROOT)/src/Make.cmd
