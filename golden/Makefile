KGO := CGO_ENABLED=0 GOOS=linux go build

.PHONY: default
default: frontend allgo

include ../go/skiaversion/skiaversion.mk

.PHONY: packages
packages:
	go build -v ./go/...

# This is the quick version triggered by the presubmit script.
.PHONY: testgo
testgo:
	go test ./go/... -short

.PHONY: bench
bench:
	go test ./go/... -bench=. -run=NONE

.PHONY: test
test: go-test js-test puppeteer-test
	true

.PHONY: go-test
go-test:
	go test ./go/...

.PHONY: js-test
js-test:
	xvfb-run --auto-servernum --server-args "-screen 0 1280x1024x24" npx karma start --single-run

.PHONY: js-test-debug
js-test-debug:
	npx karma start --no-single-run

.PHONY: puppeteer-test
puppeteer-test: clean-puppeteer-tests-output
	docker run --interactive --rm \
		--mount type=bind,source=`pwd`/..,target=/src \
		--mount type=bind,source=`pwd`/puppeteer-tests/output,target=/out \
		gcr.io/skia-public/gold-puppeteer-tests:latest \
		/src/golden/puppeteer-tests/docker/run-tests.sh

.PHONY: puppeteer-test-nodocker
puppeteer-test-nodocker: clean-puppeteer-tests-output
	# If we run "npx mocha" outside of puppeteer-tests, file
	# puppeteer-tests/.mocharc.json won't be read.
	cd puppeteer-tests; npx mocha

.PHONY: puppeteer-test-debug
puppeteer-test-debug: clean-puppeteer-tests-output
	# If we run "npx mocha" outside of puppeteer-tests, file
	# puppeteer-tests/.mocharc.json won't be read.
	cd puppeteer-tests; npx mocha --inspect-brk

.PHONY: clean-puppeteer-tests-output
clean-puppeteer-tests-output:
	rm -rf puppeteer-tests/output
	mkdir -p puppeteer-tests/output

node_modules: package.json
	npm install

.PHONY: frontend
frontend:
	cd frontend && $(MAKE) web
	npx webpack-cli --mode=production

.PHONY: frontend-debug
frontend-debug:
	cd frontend && $(MAKE) web
	npx webpack-cli --mode=development

.PHONY: serve
serve:
	npx webpack-dev-server --watch-poll --mode=development

.PHONY: allgo
allgo: skiaversion
	go install -v ./cmd/...

.PHONY: k8s-release-skiacorrectness
k8s-release-skiacorrectness: build-static-skiacorrectness frontend
	./k8s_release_skiacorrectness

.PHONY: k8s-release-diff-server
k8s-release-diff-server: build-static-diffserver
	./k8s_release_diff_server

.PHONY: k8s-release-baseline-server
k8s-release-baseline-server: build-static-baseline-server
	./k8s_release_baseline_server

.PHONY: k8s-release-ingestion
k8s-release-ingestion: build-static-ingestion
	./k8s_release_ingestion

.PHONY: k8s-release-goldpushk-healthy-server
k8s-release-goldpushk-healthy-server: build-static-goldpushk-healthy-server
	./k8s_release_goldpushk_healthy_server

.PHONY: k8s-release-goldpushk-crashing-server
k8s-release-goldpushk-crashing-server: build-static-goldpushk-crashing-server
	./k8s_release_goldpushk_crashing_server

.PHONY: build-static-skiacorrectness
build-static-skiacorrectness: skiaversion frontend
	mkdir -p ./build
	rm -f ./build/skiacorrectness_k8s
	$(KGO) -o build/skiacorrectness_k8s -a ./cmd/skiacorrectness/main.go

.PHONY: build-static-baseline-server
build-static-baseline-server: skiaversion
	mkdir -p ./build
	rm -f ./build/baseline_server_k8s
	$(KGO) -o build/baseline_server_k8s -a ./cmd/baseline_server/main.go

.PHONY: build-static-diffserver
build-static-diffserver: skiaversion
	mkdir -p ./build
	rm -f ./build/diff_server_k8s
	$(KGO) -o build/diff_server_k8s -a ./cmd/skia_diff_server/main.go

.PHONY: build-static-ingestion
build-static-ingestion: skiaversion
	mkdir -p ./build
	rm -f ./build/gold_ingestion_k8s
	$(KGO) -o build/gold_ingestion_k8s -a ./cmd/gold_ingestion/main.go

.PHONY: build-static-goldpushk-healthy-server
build-static-goldpushk-healthy-server:
	mkdir -p ./build
	rm -f ./build/goldpushk_healthy_server_k8s
	$(KGO) -o build/goldpushk_healthy_server_k8s -a ./cmd/goldpushk/testing/healthy_server/main.go

.PHONY: build-static-goldpushk-crashing-server
build-static-goldpushk-crashing-server:
	mkdir -p ./build
	rm -f ./build/goldpushk_crashing_server_k8s
	$(KGO) -o build/goldpushk_crashing_server_k8s -a ./cmd/goldpushk/testing/crashing_server/main.go

.PHONY: mocks
mocks:
	go get github.com/vektra/mockery/...

	go generate ./...

include ../webtools/webtools.mk
