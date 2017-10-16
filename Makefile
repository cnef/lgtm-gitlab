.SILENT :
.PHONY : lgtm clean fmt

TAG:=`git describe --abbrev=0 --tags`
LDFLAGS:=-X main.buildVersion=$(TAG)

all: lgtm

lgtm:
	echo "Building lgtm"
	go install -ldflags "$(LDFLAGS)"

dist-clean:
	rm -rf dist

dist: dist-clean
	mkdir -p dist/ && GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -a -tags netgo -installsuffix netgo -o dist/lgtm

release: dist
	docker build -t registry.newben.net/library/lgtm .
	docker push registry.newben.net/library/lgtm