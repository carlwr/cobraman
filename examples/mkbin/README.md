For Go projects using `cobra`, the `cobraman/mkbin` package can be used to create a companion util capable of generating the man pages for the project. By compiling the companion util to a binary, the man pages can be generated independently of `cobraman`.

In the following example,
* `boodbye` is an example Go project using `cobra`, and
* `docsgen` is a companion util for `boodby`. `docsgen` uses `cobraman/mkbin`.

Below is a walk-through of how to
1. generate the code for `boodbye`
  * in `boodbye/`, using `cobra-cli`,
2. create `docsgen`
  * in `boodbye/docsgen`
3. verify that sources are in order and without errors
4. build the binaries for `boodbye` and `docsgen`
  * as `boodbye-bin` and `docsgen-bin`

The source files resulting of 1. and 2. above are checked-in to the `cobraman` repo in the `boodbye/` dir. Since we will illustrate how to generate these sources, our first step will be to delete all generated source files, assuming the `cobraman` repo has been cloned.

`boodbye/Makefile` is able to perform any of the steps 1. - 4. above. Below, the process is illustrated both _(I)_ using shell commands and _(II)_ with `make`. To illustrate the output of commands, indented shell comments are used.

# Preparation

## Required tools

`cobra-cli` is used to generate the example Go project:
```sh
go install github.com/spf13/cobra-cli@latest
```
_Optional:_ For using the Makefile, GNU make, available on most distros, is needed. The makefile has been tested with GNU make 3.81 and 4.4.1.
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
tree
  # .
  # ├── Makefile
  # └── docsgen
  #     └── docsgen.go
```

# 1. generate the code for `boodbye`

Create `go.mod`:

```sh
# (I) with shell commands:
>go.mod cat <<EOF
module github.com/carlwr/cobraman/examples/mkbin/boodbye
go 1.23
replace github.com/carlwr/cobraman v0.0.0 => ../../..
require github.com/carlwr/cobraman v0.0.0
EOF

# or (II) with make:
make go.mod
```

Initialize the project as a `cobra` project using `cobra-cli`, then patch the generated `root.go` file to also contain a function `GetRootCmd()`, and leave only the patched file as `cmd/rootCmd.go`:

```sh

# (I) with shell commands:
cobra-cli init

> cmd/rootCmd.go  \
  cat cmd/root.go
>>cmd/rootCmd.go  \
  printf '%s\n' \
    '// Added for use by cobraman/mkbin.' \
    'func GetRootCmd() *cobra.Command {' \
    $'\treturn rootCmd' \
    '}'
rm cmd/root.go

rm LICENSE  # generated file that isn't needed

# (II) with make:
make cmd/rootCmd.go main.go

# result:
tree
  # ./
  # ├── cmd/
  # │   └── rootCmd.go
  # ├── docsgen/
  # │   └── docsgen.go
  # ├── Makefile
  # ├── go.mod
  # ├── go.sum
  # └── main.go
```

# 2. create `docsgen`

In a typical workflow, this is the point where the companion tool (named `docsgen` in this example) would be added to the project. For this example, we had the `docsgen/docsgen.go` file in place from the beginning.

# 3. verify that sources are in order and without errors

```sh
go mod tidy

go vet main.go
go vet cmd/*.go
go vet docsgen/docsgen.go
```

# 4. build the binaries for `boodbye` and `docsgen`

```sh

# (I) with shell commands:
go mod tidy
go build -o main-bin main.go
go build -o docsgen/docsgen-bin docsgen/docsgen.go

# (II) with make:
make binaries

# result:
tree --prune -P '*-bin'
  # ./
  # ├── docsgen/
  # │   └── docsgen-bin*
  # └── main-bin*

```
