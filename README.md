<div align="center">
<p>
    <img width="300" src="https://raw.githubusercontent.com/fract-lang/resources/main/logo/fract.svg?sanitize=true">
</p>
<h1>The Fract Programming Language</h1>
Fast, efficient, reliable, safe and simple. <br>
Designed for powerful scripting programming.
    
<strong>

[Website](https://fract-lang.github.io/website/) |
[Docs](https://fract-lang.github.io/website/pages/docs/docs.html) |
[A Tour of Fract](https://fract-lang.github.io/website/pages/tour.html) |
[Samples](https://fract-lang.github.io/website/pages/samples.html) |
[Contributing](#contributing)
</strong>
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
```go
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

## Object Oriented Programming
### Structures
```v
struct People {
    Name
    Surname
}

p := People('Tony', 'Stark')
println(p) // {Name:Tony Surname:Stark}
```
### Classes
```csharp
class Employee {
  var (
    Name    = ''
    Surname = ''
    Age     = 0
    Salary  = 0
  )

  // Constructor.
  fn Employee(name, surname, age, salary=9000) {
    this.Name    = name.trim()
    this.Surname = surname.trim()
    this.Age     = age
    this.Salary  = salary
  }

  fn FullName() {
    ret this.Name + ' ' + this.Surname
  }

  fn InfoString() {
    ret (
      'Name: '     + this.Name +
      ' Surname: ' + this.Surname +
      ' Age: '     + string(this.Age) +
      ' Salary: '  + string(this.Salary)
    )
  }
}

e := Employee('Daniel', 'Garry', 44, 12550)
println(e.FullName())   // Daniel Garry
println(e.InfoString()) // Name: Daniel Surname: Garry Age: 44 Salary: 12550
```

## Interactive Shell
Fract has an interactive shell. You can try quickly without writing the codes to the file. <br>
Exampe on Manjaro Linux;
```shell
$ ./fract
Fract 0.0.1 (c) MIT License.
Fract Developer Team.

>> var name = input('Username: ')
Username: fract
>> var pass = input('Password: ')
Password: root
>> if name == 'fract' && pass = 'root' {
 |   println('success!')
 | } else {
 |   println('failed!')
 | }
Success!
>> exit(0)
```

## How to Compile
There are scripts prepared for compiling of Fract. <br>
These scripts are written to run from the home directory. <br>

``brun`` scripts used for compile and execute if compiling is successful. <br>
`build` scripts used for compile. <br>

Fract is compiled from a single file. <br>
You can write a custom script yourself or choose to compile/run it directly from go compiler. <br>
For compile: ``go build cmd/main.go`` <br>
For run: ``go run cmd/main.go``

## License
Fract is distributed under the terms of the MIT license. <br>
[See license details.](https://fract-lang.github.io/website/pages/license.html)

<h2 id="contributing">Contributing</h2>
Thanks for you want contributing to Fract!
<br>
The Fract project use issues for only bug reports and proposals.
<br><br>
To contribute, please read the contribution guidelines from <a href="https://fract-lang.github.io/website/pages/contributor_guide.html">here</a>.
<br><br>
To discussions and questions, please use <a href="https://github.com/fract-lang/fract/discussions">discussions</a>.
