For Go projects using `cobra`, the `cobraman/mkbin` package can be used to create a companion util capable of generating the man pages for the project. By compiling the companion util to a binary, the man pages can be generated independently of `cobraman`.

In the following example,
* `boodbye` is our example Go project (that uses `cobra`), and
* `docsgen` is the companion util for `boodby` that we will create (that uses `cobraman/mkbin`).

Below is a walk-through of how to
1. generate the code for `boodbye`
  * in `boodbye/`, using `cobra-cli`,
2. create `docsgen`
  * in `boodbye/docsgen`
3. verify that sources are in order and without errors
4. build the binaries for `boodbye` and `docsgen`
  * as `boodbye-bin` and `docsgen/docsgen-bin`
5. test and use the binaries

The source files resulting of 1. and 2. above are checked-in to the `cobraman` repo in the `examples/mkbin/boodbye/` dir. Since we will illustrate how to generate these sources, our first step will be to delete them, assuming the `cobraman` repo has been cloned.

The makefile `boodbye/Makefile` is able to perform any of the individual steps 1. - 4. above. To perform all steps at once the command `make clean && make binaries` can be used.

In what follows, terminal commands are provided both _(I)_ as explicit commands and _(II)_ as `make` commands. For expected terminal output, indented shell comments are used.

# Preparation

## Required tools

`cobra-cli` is used to generate the example Go project:
```sh
go install github.com/spf13/cobra-cli@latest
```

_Optional:_ Using the makefile requires _GNU make_. The makefile has been tested with GNU make 3.81 and 4.4.1.
```sh
make --version | head -n1
  # GNU Make 4.4.1
```

## Prepare the directory

Clone the repo and change to the `boodbye/` dir:
```sh
git clone https://github.com/carlwr/cobraman
cd cobraman/examples/mkbin/boodbye
```

Remove all of the files but `Makefile` and `docsgen/docsgen.go`, as we are just about to generate them anew:

```sh
# (I) with shell commands:
rm -rf cmd/
rm -f  main.go go.mod go.sum

# or (II) with make:
make clean

# result:
tree -F
  # .
  # ├── Makefile
  # └── docsgen
  #     └── docsgen.go
```

# Steps

## 1. generate the code for `boodbye`

Create `go.mod`:

```sh
# (I) with shell commands:
touch go.mod
go mod edit \
  -module   github.com/carlwr/cobraman/examples/mkbin/boodbye \
  -go       1.23 \
  -replace  github.com/carlwr/cobraman@v0.0.0=../../.. \
  -require  github.com/carlwr/cobraman@v0.0.0

# or (II) with make:
make go.mod
```

Initialize the project using `cobra-cli`, then patch the generated `root.go` file to also contain a function `GetRootCmd()`:

```sh

# (I) with shell commands:
cobra-cli init
>> cmd/root.go \
   printf '%s\n' 'func GetRootCmd() *cobra.Command { return rootCmd }'
rm LICENSE  # generated file that isn't needed

# (II) with make:
make cmd/root.go main.go

# result:
tree -F
  # ./
  # ├── cmd/
  # │   └── root.go
  # ├── docsgen/
  # │   └── docsgen.go
  # ├── Makefile
  # ├── go.mod
  # ├── go.sum
  # └── main.go

# (go.sum was created as an effect of running `cobra-cli init`; which also updated go.mod with the cobra dependency.)
```

Use `cobra-cli` to add two commands:

```sh

# (I) with shell commands:
cobra-cli add hello
cobra-cli add boodbye

# (II) with make:
make cmd/hello.go cmd/boodbye.go

# result:
tree -F
  # ./
  # ├── cmd/
  # │   ├── boodbye.go
  # │   ├── hello.go
  # │   └── root.go
  # ├── docsgen/
  # │   └── docsgen.go
  # ├── Makefile
  # ├── go.mod
  # ├── go.sum
  # └── main.go

```


## 2. create `docsgen`

In a typical workflow, this is the point where the companion tool (named `docsgen` in this example) would be added to the project. For this example, we had the `docsgen/docsgen.go` file in place from the beginning.

## 3. verify that sources are in order and without errors

```sh
go mod tidy

go vet main.go
go vet cmd/*.go
go vet docsgen/docsgen.go
```

## 4. build the binaries for `boodbye` and `docsgen`

```sh

# (I) with shell commands:
go mod tidy
go build -o main-bin            main.go
go build -o docsgen/docsgen-bin docsgen/docsgen.go

# (II) with make:
make binaries

# result:
tree -F
  # ./
  # ├── cmd/
  # │   ├── boodbye.go
  # │   ├── hello.go
  # │   └── root.go
  # ├── docsgen/
  # │   ├── docsgen-bin*
  # │   └── docsgen.go
  # ├── Makefile
  # ├── go.mod
  # ├── go.sum
  # ├── main-bin*
  # └── main.go

```

## 5. test and use the binaries

Test the `boodbye` program itself:

```sh

./main-bin hello
  # hello called

./main-bin boodbye
  # boodbye called

```

Test the `boodbye`s companion util `docsgen` that we just created:

```sh

# view the help:
./docsgen/docsgen-bin --help
  # Generate documentation, etc.
  # 
  # Usage:
  #   docsgen [command]
  # 
  # Available Commands:
  #   completion             Generate the autocompletion script for the specified shell
  #   generate-auto-complete Generate bash auto complete script
  #   generate-markdown      Generate docs with the markdown template
  #   generate-mdoc          Generate docs with the mdoc template
  #   generate-troff         Generate docs with the troff template
  #   help                   Help about any command
  # 
  # Flags:
  #       --directory string   Directory to install generated files (default ".")
  #   -h, --help               help for docsgen
  # 
  # Use "docsgen [command] --help" for more information about a command.


# generate markdown manual for `boodbye`:
mkdir -p docs/markdown
./docsgen/docsgen-bin generate-markdown --directory docs/markdown

# result:
tree -F docs
  # docs/
  # └── markdown/
  #     ├── boodbye.md
  #     ├── boodbye_boodbye.md
  #     └── boodbye_hello.md

```

---

# Notes

The example application is named "boodbye" in tribute to [ee194c2](https://github.com/rayjohnson/cobraman/blob/ee194c2025975261ec1ba172d2f68fabca6a819d/example/cmd/boodbye.go).