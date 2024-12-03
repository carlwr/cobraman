
```sh

# the below is supposed to be run in:
print -D $PWD
  # ~dev/go/cobraman

# (commands below were actually executed in a temp directory; prefix of path outputs below replaced with
#     [..]/cobraman/
#  with intention
#     [..]/cobraman/ <=> ~dev/go/cobraman

cd examples/mkbin


mkdir boodbye && ( cd boodbye && go mod init pkgpath/boodbye )
  # go: creating new go.mod: module pkgpath/boodbye

tre
  # ./
  # └── boodbye/
  #     └── go.mod

( cd boodbye && cobra-cli init )
  # Your Cobra application is ready at
  # [..]/cobraman/examples/mkbin/boodbye

mkbin tre
  # ./
  # └── boodbye/
  #     ├── cmd/
  #     │   └── root.go
  #     ├── LICENSE
  #     ├── go.mod
  #     ├── go.sum
  #     └── main.go

( cd boodbye && cobra-cli add hello )
  # hello created at [..]/cobraman/examples/mkbin/boodbye

tre
  # ./
  # └── boodbye/
  #     ├── cmd/
  #     │   ├── hello.go
  #     │   └── root.go
  #     ├── LICENSE
  #     ├── go.mod
  #     ├── go.sum
  #     └── main.go

rg -tgo '^package ' | column -t -s: | sort -k3
  # boodbye/cmd/hello.go  package cmd
  # boodbye/cmd/root.go   package cmd
  # boodbye/main.go       package main


# after `go mod init` and `cobra-cli ..`s above, the intention is to instruct the user to run:
mkdir boodbye/tools
cp example_main.go boodbye/tools/main.go


```




recording current file tree of actual dir `~dev/go/cobraman`:
```sh

print -D $PWD      # ~dev/go/cobraman
cd examples/mkbin
tre
  # ./
  # ├── boodbye/
  # │   ├── cmd/
  # │   │   ├── boodbye.go
  # │   │   ├── hello.go
  # │   │   └── root.go
  # │   └── tools/
  # │       ├── docutil*
  # │       └── main.go
  # ├── LICENSE
  # ├── README-todo-mod.md
  # ├── README.md
  # └── example_main.go

```