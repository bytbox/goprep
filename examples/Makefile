include ${GOROOT}/src/Make.inc

TARGS = id nocomments goforpp

all: ${TARGS}

clean:
	rm *.${O} ${TARGS}

id: id.${O}
	${LD} -o $@ id.${O}

id.${O}: id.go
	${GC} id.go

nocomments: nocomments.${O}
	${LD} -o $@ nocomments.${O}

nocomments.${O}: nocomments.go
	${GC} nocomments.go

goforpp: goforpp.${O}
	${LD} -o $@ goforpp.${O}

goforpp.${O}: goforpp.go
	${GC} goforpp.go
