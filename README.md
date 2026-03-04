# sfuzz

Simple blackbox fuzzer to harden the validation layer of JSON API endpoints. 

## Features

Compared to other black box API fuzzers this one was designed to have a new particular 
set of features. 

- simple FUZZ file format to capture all fuzzing requests
- autogeneration of FUZZ files from any Open API specification (> 3.x)
- converge quicker via embedded FUZZ types and generation of happy path values
- run simultaneously on any number of requests

Importantly, as a black box fuzzer it does not mutate based on static/dynamic source code feedback.

## Resilient APIs

APIs are a front to many products and businesses. We like:

- more 4xx business responses and less unexpected 5xx status codes
- less noise in logs and alerting 
- less out of bonds and overflows errors
- continuously stress testing of our business validation layer

## LLM Friendly

`sfuzz` is designed so that LLMs can also discover shortcomings in APIs:

- concise, structured and meaningful console outputs
- clear text and no colored console output
- separate steps for fuzz file generation and actual fuzzing

_Coming soon_: A skill directory.

## Install

```console
go install github.com/simcap/sfuzz/cmd/sfuzz
```

## How it works?

![sfuzz diagram](./sfuzz.png)
