.PHONY: test test-iacommon test-iavm test-ialang test-release-0.0.5

test: test-iacommon test-iavm test-ialang

test-iacommon:
	go test ./iacommon/...

test-iavm:
	go test ./iavm/...

test-ialang:
	go test ./ialang/...

test-release-0.0.5:
	go test ./iacommon/... ./iavm/... ./ialang/...
