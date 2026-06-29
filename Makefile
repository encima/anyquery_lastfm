
files := $(filter-out %_test.go,$(wildcard *.go))

all: $(files)
	go build -o anyquery_lastfm.out $(files)

prod: $(files)
	go build -o anyquery_lastfm.out -ldflags "-s -w" $(files)

release: prod
	goreleaser build -f .goreleaser.yaml --clean --snapshot

clean:
	rm -f anyquery_lastfm.out

.PHONY: all clean
