.PHONY: build
build: web ingest_json_validator
	go install -v ./go/...

buildk: web
	CGO_ENABLED=0 GOOS=linux go install -a ./go/skiaperf
	./build_docker_release

build-perf-ingest:
	CGO_ENABLED=0 GOOS=linux go install -a ./go/perf-ingest
	./build_perf_ingest_release

push-perf-ingest: build-perf-ingest
	pushk perf-ingest

pushk: buildk
	pushk skiaperf

.PHONY: perf-force-ingest
	go install -v ./go/perf-force-ingest

ingest_json_validator:
	go install -v ./go/ingest_json_validator

start_datastore_emulator:
	 docker run -ti -p 8008:8008 google/cloud-sdk:latest gcloud beta emulators datastore start --no-store-on-disk --project test-project --host-port 0.0.0.0:8008

package-lock.json: package.json
	npm ci
	touch package-lock.json

.PHONY: web
web: package-lock.json
	npx webpack-cli --mode=production

watch: package-lock.json
	npx webpack --mode=development --watch

.PHONY: testjs
testjs:
	# Run the generated tests just once under Xvfb.
	xvfb-run --auto-servernum --server-args "-screen 0 1280x1024x24" npx karma start --single-run

.PHONY: test
test: testgo testjs

.PHONY: testgo
testgo:
	go test ./go/...

.PHONY: debug
debug:
	npx webpack-cli --mode=development

.PHONY: all # For ../Makefile 'all' target.
all:
	go install -v ./go/...
	# Go back to npm ci after npm is upgraded on the bot.
	npm i
	npx webpack-cli --mode=production

start_bigtable_emulator:
	docker run -ti -p 8000:8000 google/cloud-sdk:latest gcloud beta emulators bigtable start \
		  --project test-project --host-port 0.0.0.0:8000