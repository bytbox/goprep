TARGS = id

all: ${TARGS}

clean:
	rm *.6 ${TARGS}

id: id.6
	6l -o $@ id.6

id.6: id.go
	6g id.go
