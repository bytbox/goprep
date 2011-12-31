include ${GOROOT}/src/Make.inc

TARGS = id nocomments

all: ${TARGS}

clean:
	rm *.${O} ${TARGS}

id: id.${O}
	${LD} -o $@ id.6

id.${O}: id.go
	${GC} id.go

nocomments: nocomments.${O}
	${LD} -o $@ nocomments.6

nocomments.${O}: nocomments.go
	${GC} nocomments.go
