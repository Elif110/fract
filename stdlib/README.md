# stdlib

Standard library directory. <br>
The directories in this directory, accept as package. <br>
The files in this directory, accept as local package files. <br>

## Important

Local package files is should not open another packages from standard library.

Example;
```go
package std

struct example {
    a
    b
}
```

This struct define above can use from anything fract source code without any import operation.
