<div align="center">
<p>
    <img width="300" src="https://raw.githubusercontent.com/fract-lang/resources/main/logo/fract.svg?sanitize=true">
</p>
<h1>The Fract Programming Language</h1>

[Website](https://fract-lang.github.io/website/) |
[Docs](https://fract-lang.github.io/website/pages/docs/docs.html) |
[A Tour of Fract](https://fract-lang.github.io/website/pages/tour.html) |
[Samples](https://fract-lang.github.io/website/pages/samples.html) |
[Contributing](#contributing)

</div>

## Key features
+ Simplicity: easy to learn, can be learned in less than an hour
+ Safety: no null, no undefined values, immutability by default
+ Unreachable codes are not included in debugging
+ Efficient and performance
+ Object Oriented Programming
+ Deferred calls
+ Language level concurrency
+ Pragmas

## What look like Fract code?

```v
package main

fn fib(a, b) {
    val := a + b
    println(val)
    if val < 1000 {
        fib(b, val)
    }
}

fib(0, 1)
```

## Interactive Shell
<img src="https://github.com/fract-lang/resources/blob/main/preview/fract_cli.gif?raw=true">

## How to compile?
Fract is written in Go. <br>
Run one of the scripts ``scripts/brun.bat`` or ``scripts/brun.sh`` to compile. <br>
Also can be write manually: ``go build cmd/main.go``

<h2 id="contributing">Contributing</h2>
Thanks for you want contributing to Fract!
<br>
The Fract project use issues for only bug reports and proposals.
<br><br>
To contribute, please read the contribution guidelines from <a href="https://fract-lang.github.io/website/pages/contributor_guide.html">here</a>.
<br><br>
To discussions and questions, please use <a href="https://github.com/fract-lang/fract/discussions">discussions</a>.
