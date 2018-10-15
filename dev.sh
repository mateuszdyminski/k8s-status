#!/bin/bash

usage() {
	cat <<EOF
Usage: $(basename $0) <command>

Wrappers around core binaries:
    run                    Runs k8s-status locally.
    build                  Builds backend - static binary is in 'bin' directory.
    docker-build           Builds docker image based on existing go binary.
    docker-push            Pushes docker image to dockerhub.
    docker-all             Runs 'build', 'build-ui', 'docker-build' and 'docker-push' commands.
EOF
	exit 1
}

APP_VERSION=$(git describe --always)

build() {
	GITREPO='github.com/mateuszdyminski/k8s-status'
	APP_NAME='k8s-status'
	LAST_COMMIT_USER="$(tr -d '[:space:]' <<<"$(git log -1 --format=%cn)<$(git log -1 --format=%ce)>")"
	LAST_COMMIT_HASH=$(git log -1 --format=%H)
	LAST_COMMIT_TIME=$(git log -1 --format=%cd --date=format:'%Y-%m-%d_%H:%M:%S')

	LDFLAGS="-s -w -X $GITREPO/pkg/version.AppName=$APP_NAME -X $GITREPO/pkg/version.AppVersion=$APP_VERSION -X $GITREPO/pkg/version.LastCommitTime=$LAST_COMMIT_TIME -X $GITREPO/pkg/version.LastCommitHash=$LAST_COMMIT_HASH -X $GITREPO/pkg/version.LastCommitUser=$LAST_COMMIT_USER -X $GITREPO/pkg/version.BuildTime=$(date -u +%Y-%m-%d_%H:%M:%S)"
	
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o build/k8status -a -tags netgo main.go
}

buildDocker() {
	docker build -t mateuszdyminski/k8s-status:latest -t mateuszdyminski/k8s-status:$APP_VERSION .
}

pushDocker() {
	docker push mateuszdyminski/k8s-status:$APP_VERSION
	docker push mateuszdyminski/k8s-status:latest

	echo "Application pushed to repo mateuszdyminski/k8s-status:$APP_VERSION" 
}

run() {
	K8STATUS_HTTPPORT=8090 K8STATUS_GRACEFULSHUTDOWNTIMEOUT=5 K8STATUS_GRACEFULSHUTDOWNEXTRASLEEP=0 \
	K8STATUS_DEBUG=true K8STATUS_KUBECONFIGPATH=~/.kube/config go run main.go 
}

CMD="$1"
SUBCMD="$2"
shift
case "$CMD" in
	run)
		run
	;;
	build)
		build
	;;
	docker-build)
		buildDocker
	;;
    docker-push)
		pushDocker
	;;
	docker-all)
		build
		buildDocker
		pushDocker
	;;
	*)
		usage
	;;
esac
