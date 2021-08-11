.SILENT :
.PHONY : test dist release

TAG:=`git describe --abbrev=0 --tags`
LDFLAGS:=-X main.buildVersion=$(TAG)
REGISTRY:=docker.io

test:
	LGTM_GITLAB_URL=http://192.168.1.11:8000 \
	LGTM_COUNT=1 \
	LGTM_TOKEN=zXHTLu1azQ1qxQ3xkXmu \
	LGTM_DB_PATH=./data.o \
	go run ./

dist:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/lgtm

release: dist
	docker buildx build --platform=linux/amd64 --push -t $(REGISTRY)/library/lgtm-gitlab .

docker-run:
	docker run --name lgtm -itd -p 8989:8989 \
	-e LGTM_GITLAB_URL=http://192.168.1.11:8000 \
	-e LGTM_COUNT=1 \
	-e LGTM_TOKEN=zXHTLu1azQ1qxQ3xkXmu \
	$(REGISTRY)/library/lgtm-gitlab
