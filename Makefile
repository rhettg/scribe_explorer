include $(GOROOT)/src/Make.inc

TARG=rangerweb
GOFILES=\
	data_stream.go\
	get_deep.go\
	rangerweb.go\
	filter.go\
	aggregate.go\
	json_io.go

include $(GOROOT)/src/Make.cmd
