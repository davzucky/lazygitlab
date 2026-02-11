.PHONY: build run test lint fmt tidy tui-validate pre-commit hooks commit-check ci

# Deprecated: prefer running recipes from justfile (`just <recipe>`).
# Keep Make targets as temporary compatibility wrappers.
build:
	just build

run:
	just run

test:
	just test

lint:
	just lint

fmt:
	just fmt

tidy:
	just tidy

tui-validate:
	just tui-validate

pre-commit:
	just pre-commit

hooks:
	just hooks

commit-check:
	just commit-check

ci:
	just ci
