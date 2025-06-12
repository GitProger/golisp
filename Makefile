build = _build

build: clean
	mkdir $(build)
	cd cmd/repl && go build -o ../../$(build)/repl
run:
	./$(build)/repl
runrl:
	rlwrap ./$(build)/repl
clean:
	rm -r $(build)/ || true

all: clean build run
