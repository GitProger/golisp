build = _build

build: clean
	mkdir $(build)
	cd cmd/repl && go build -o ../../$(build)/repl
run:
	./$(build)/repl
runrl:
	rlwrap ./$(build)/repl
clean:
	[ -e $(build) ] && rm -r $(build)/

all: clean build run
