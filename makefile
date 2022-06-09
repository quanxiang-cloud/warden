name=warden
version=alpha
devHost=192.168.200.20
devUser=ubuntu

repository=lowcode
dockerHost=dockerhub.qingcloud.com

env:
#-- open go mod vendor --
	go mod vendor
dev: linux
	./dev_auto.sh
linux:
	GOOS=linux GOOSARCH=amd64 go build -o $(name) ./cmd/.
docker-test: env
	docker build -t $(dockerHost)/$(repository)/$(name):$(version) .
	docker push  $(dockerHost)/$(repository)/$(name):$(version)
