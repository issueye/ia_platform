.PHONY: test test-iacommon test-iavm test-ialang

test: test-iacommon test-iavm test-ialang

test-iacommon:
	go test ./iacommon/...

test-iavm:
	go test ./iavm/...

test-ialang:
	go test ./ialang/...
