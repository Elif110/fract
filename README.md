<div align="center">
<p>
    <img width="300" src="https://raw.githubusercontent.com/fract-lang/resources/main/logo/fract.svg?sanitize=true">
</p>
<h1>The Fract Programming Language</h1>
Fast, efficient, reliable, safe and simple. <br>
Designed for powerful scripting programming.
    
<strong>

[Website](https://fract-lang.github.io/website/) |
[Documentations](https://fract-lang.github.io/website/pages/docs/docs.html) |
[A Tour of Fract](https://fract-lang.github.io/website/pages/tour.html) |
[Samples](https://fract-lang.github.io/website/pages/samples.html)
</strong>
</div>

## Table of Contents
<div class="toc">
  <ul>
    <li>
      <a href="#overview">Overview</a>
      <ul>
        <li><a href="#key_features">Key Features</a></li>
      </ul>
    </li>
    <li><a href="#future_changes">Future Changes</a></li>
    <li>
      <a href="#object_oriented_programming">Object Oriented Programming</a>
      <ul>
        <li><a href="#structs">Structs</a></li>
        <li><a href="#classes">Classes</a></li>
      </ul>
    </li>
    <li><a href="#interactive_shell">Interactive Shell</a></li>
    <li><a href="#how_to_run_fract_code">How to run Fract Code</a></li>
    <li><a href="#how_to_compile">How to Compile</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
  </ul>
</div>

<h2>Overview</h2>
Fract aims to be a powerful interpreted programming language. <br>
It is focused on being a good choice to a powerful and modern scripting languages. <br>
In addition, it gives importance to simplicity and readability. Fract code looks pretty plain and readable.
<br><br>
<strong>Fibonacci series;</strong>

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

<h3 id="key_features">Key features</h3>

+ Simplicity: easy to learn, can be learned in less than an hour
+ Safety: no null, no undefined values, immutability by default
+ Unreachable codes are not included in debugging
+ Efficient and performance
+ Object Oriented Programming
+ Deferred calls
+ Language level concurrency
+ Pragmas

<h2 id="future_changes">Future Changes</h2>

Although Fract is still in early development, future codes will not be much different from what they are now. Since there is no main version yet, it may undergo syntax changes, but these will not be major changes. Fract is relatively fast and stable. As his library expands, so will his abilities. However, it will continue to be a simple and plain language in the future.

<h2 id="object_oriented_programming">Object Oriented Programming</h2>

Fract has the advantages of the object-oriented programming approach and aims to do so without violating the goals of readability and simplicity.
Has adopted two structures for this namely ``struct`` and ``class``.

<h3 id="structs">Structs</h3>

Structures are a field collection that can contain only fields. It cannot have default values and constructor methods. <br>
Constructor methods are automatically defined by Fract.

```go
package main

struct People {
    Name
    Surname
}

p := People('Tony', 'Stark')
println(p) // {Name:Tony Surname:Stark}
```

<h3 id="classes">Classes</h3>

Classes can contain fields, methods and have a constructor method privately.

```csharp
package main

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

<h2 id="interactive_shell">Interactive Shell</h2>

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

<h2 id="how_to_run_fract_code">How to run Fract code</h2>

Example on Manjaro Linux; <br>
Fract file: ``main.fra``
```go
package main

println("Hello, World!")
```
Run code:
```
$ ./fract main.fra
Hello, World!
$ 
```

<h2 id="how_to_compile">How to Compile</h2>

There are scripts prepared for compiling of Fract. <br>
These scripts are written to run from the home directory. <br>

``brun`` scripts used for compile and execute if compiling is successful. <br>
`build` scripts used for compile. <br>

Fract is compiled from a single file. <br>
You can write a custom script yourself or choose to compile/run it directly from go compiler. <br>
For compile: ``go build cmd/main.go`` <br>
For run: ``go run cmd/main.go``

<h2 id="contributing">Contributing</h2>

Thanks for you want contributing to Fract!
<br><br>
The Fract project use issues for only bug reports and proposals. <br>
To contribute, please read the contribution guidelines from <a href="https://fract-lang.github.io/website/pages/contributor_guide.html">here</a>. <br>
To discussions and questions, please use <a href="https://github.com/fract-lang/fract/discussions">discussions</a>.
<br><br>
All contributions to Fract, no matter how small or large, are welcome. <br>
From a simple typo correction to a contribution to the code, all contributions are welcome and appreciated. <br>
Before you start contributing, you should familiarize yourself with the following repository structure; <br>

+ ``cmd/`` main and compile files.
+ ``functions/`` built-In functions.
+ ``lex/`` lexer.
+ ``oop/`` object oriented programming infrastructure and all value, object components.
+ ``parser/`` interpreter.
+ ``pkg/`` utility packages.
+ ``samples/`` sample codes of Fract.
+ ``scripts/`` the build and run, build and other all batch, bash or another scripts.
+ ``stdlib/`` the standard library of Fract.
+ ``tests/`` contains categorized tests for the interpreter and standard library.

<h2 id="license">License</h2>

Fract and standard library is distributed under the terms of the MIT license. <br>
[See license details.](https://fract-lang.github.io/website/pages/license.html)
