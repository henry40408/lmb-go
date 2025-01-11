# DEPRECATION NOTICE

I'm back maintaining the [Rust version](https://github.com/henry40408/lmb). The Lua performance in Go version is not as good as Rust, here's a simple (very unscientific and could be affected by many factors) benchmark report. This repository is archived just to prove that I have written Go before.

## Hardware

* AMD Ryzen 5 5600X 6-Core Processor

## Go

* commit [161ae67](https://github.com/henry40408/lmb-go/commit/161ae67)

```
goos: linux
goarch: amd64
pkg: github.com/henry40408/lmb/internal/eval_context
cpu: AMD Ryzen 5 5600X 6-Core Processor
BenchmarkCompile-12               179800              6654 ns/op
BenchmarkEvalCompiled-12           20778             57368 ns/op
BenchmarkEvalConcurrency-12         8607            121437 ns/op
BenchmarkEvalScript-12              8739            122985 ns/op
```

## Rust

* commit [95c6125](https://github.com/henry40408/lmb/commit/95c6125)

```
running 12 tests
test lmb_default_store    ... bench:         660 ns/iter (+/- 51)
test lmb_evaluate         ... bench:         662 ns/iter (+/- 33)
test lmb_no_store         ... bench:         666 ns/iter (+/- 31)
test lmb_read_all         ... bench:       1,188 ns/iter (+/- 216)
test lmb_read_line        ... bench:       1,223 ns/iter (+/- 48)
test lmb_read_number      ... bench:       1,100 ns/iter (+/- 72)
test lmb_read_unicode     ... bench:       1,402 ns/iter (+/- 186)
test lmb_update           ... bench:      10,733 ns/iter (+/- 701)
test mlua_call            ... bench:          45 ns/iter (+/- 5)
test mlua_eval            ... bench:       3,962 ns/iter (+/- 111)
test mlua_sandbox_eval    ... bench:       4,026 ns/iter (+/- 3,746)
test read_from_buf_reader ... bench:          23 ns/iter (+/- 1)

test result: ok. 0 passed; 0 failed; 0 ignored; 12 measured
```

---

# lmb

> Lua Function Runner

## Features

- Single binary: Easy to deploy
- Evaluate a function to get a result: In-place computation or one-shot script
- Handle HTTP requests with a script: Long-running HTTP server

## Installation

Install with `go install`:

```sh
go install github.com/henry40408/lmb@latest
lmb --version
```

## Usage

```sh
$ echo 'print("hello, world")' >> hello.lua
$ lmb eval --file hello.lua
hello, world
```

## License

MIT
