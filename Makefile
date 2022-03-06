
GC ?= gccgo
objects = \
	src/main.go \
	src/cell.go \
	src/table.go \
	src/expression.go \
	src/evaluate.go

.PHONY: all
all: gocell

gocell: $(objects)
	$(GC) $(objects) -o gocell

