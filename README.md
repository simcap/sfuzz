# sfuzz

Simple fuzzer (... and not Simon's fuzzer) to harden your JSON APIs. This fuzzer is designed
with new things to converge more quickly to meaningful findings and create less noise:

- introduces FUZZ types in order to generate values and fuzz with better context
- leverages transparently API spec and other means to have meaningful happy path values during fuzzing permutations
- works at once in one launch on specified targets and fuzz placeholders  

More resilient APIs means:

- more 4xx business responses and less unexpected 5xx status codes
- less noise in logs and alerting 
- less out of bonds and overflows errors

Friendly for LLM usage so that your favorite friend (Claude, Codex, etc.)
can leverage it on its own:

- concise, structured and meaningful console outputs
- clear text and no colored console output
- separate steps for fuzz file generation and actual fuzzing

## Install

```console
go install github.com/simcap/sfuzz
```

## How it works?
