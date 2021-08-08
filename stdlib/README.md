# stdlib

Standard library directory. <br>
The directories in this directory, accept as package. <br>
The files in this directory, accept as local package files. <br>

## Important

Local package files is must not open another packages from standard library.

Local package files imports directy to any Fract source code and can use without package name.

For Example;
```go
package std

struct example {
    a
    b
}
```

This struct define above can use from anything fract source code without any import operation.

Like;
```go
package main

println(example('A', 'B'))
```
