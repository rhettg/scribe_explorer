include $(GOROOT)/src/Make.inc

TARG=rangerweb
GOFILES=\
	rangerweb.go\
	data_stream.go\
	json_io.go\
	parse.go\
	get_deep.go\
	filter.go\
	aggregate.go\
	window.go

include $(GOROOT)/src/Make.cmd
