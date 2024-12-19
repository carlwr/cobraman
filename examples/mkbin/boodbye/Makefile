MAKEFLAGS     := --no-builtin-rules         \
                 --warn-undefined-variables \
                 --no-print-directory


# ------------------

module-path := github.com/carlwr/cobraman
go-version  := 1.23


# ------------------

t := .aux

buildable-go :=   \
	cmd/root.go     \
	cmd/boodbye.go  \
	cmd/hello.go    \
	main.go

main-bin    := main-bin
docsgen-bin := docsgen/docsgen-bin

bins :=     \
	$(main-bin)   \
	$(docsgen-bin)


# ------------------

all     : $(buildable-go) tidy
binaries: $(bins)
.PHONY  : all binaries

.DEFAULT_GOAL := all


# ------------------

$(bins): $(buildable-go) tidy

$(bins):\
%-bin  : %.go
	go build -o $@ $<
	@echo

$(buildable-go): \
%.go           : $t/%.unfmt
	$(mkdir)
	$(gofmt) $<  >$@
	@echo

%.unfmt: %.unappended %.append
	cat $^  >$@
	@echo

$t/main.unfmt $t/cmd/root.unappended &: | go.mod
	@mkdir -p $t/cmd
	cobra-cli init  $t
	  @# must use path argument, wrong import path if cd-ing
	mv $t/cmd/root.go  $t/cmd/root.unappended
	mv $t/main.go      $t/main.unfmt
	@rm -rf $t/LICENSE
	@echo

$t/cmd/hello.unfmt    \
$t/cmd/boodbye.unfmt: \
$t/cmd/%.unfmt      : | go.mod
	$(mkdir)
	$(cobra) add $(*F)
	mv $t/cmd/$(*F).go  $t/cmd/$(*F).unfmt
	@echo

go.mod:
	>$@ echo
	go mod edit \
	  -module   $(module-path)/examples/mkbin/boodbye \
	  -go       1.23 \
	  -replace  $(module-path)@v0.0.0=../../.. \
	  -require  $(module-path)@v0.0.0
	@echo

%/root.append:
	$(mkdir)
	>$@ printf '%s\n' \
	      '// Added for use by cobraman/mkbin.'  \
	      'func GetRootCmd() *cobra.Command { return rootCmd }'
	@echo

.PHONY: \
tidy
tidy: main.go cmd/root.go | go.mod
	go mod tidy
	@echo

define __

  notes
	-----

  whenever `go.mod` is a prereq, it is only ever an _order-only prereq_
    * i.e. `[..] : [..] | go.mod`
    * since: go.mod is updated _in-place_ _spuriously_ by `cobra-cli` invocations, and when we invoke `go mod tidy`
    * -> it's timestamps won't align with the actions of this Makefile
    * -> care about this files existance, but not about its timestamp
  
	go.sum is ignored by this Makefile
    * since: it will not only be created by some of the commands, but is also sometimes updated in-place
    * -> its timestamps, and possibly its existance, won't align well with the actions of this Makefile
    * further, there is no need to have `make` track neither the existance nor the timestamps of this file

endef


.PHONY:\
clean
clean: clean-binaries clean-aux
	rm -rf cmd/ main.go LICENSE
	rm -f  go.mod go.sum
	@echo

.PHONY:\
clean-binaries
clean-binaries:
	rm -rf $(bins)

.PHONY:\
clean-aux
clean-aux:
	rm -rf $t


# --------------------------- tools --------------------------- #

mkdir = @mkdir -p $(@D)

# have-cobra != command -v cobra-cli >/dev/null && echo 1
have-cobra := $(shell command -v cobra-cli >/dev/null && echo 1)
#have-gofmt != command -v gofmt     >/dev/null && echo 1
have-gofmt := $(shell command -v gofmt     >/dev/null && echo 1)

# $(warning hc  _$(have-cobra).)
# $(warning hc2 _$(have-cobra2).)
# 
# ifeq 'a' 'b'
#   $(error a b)
# else
#   $(warning OK)
# endif


ifeq '$(have-cobra)' '1'
  cobra := cd $t && cobra-cli
else
  # $(warning 2 _$(have-cobra).)
  $(error need cobra-cli: `go install github.com/spf13/cobra-cli@latest`)
endif

# clean:
# 	echo all:$(have-cobra).

ifeq '$(have-gofmt)' '1'
  gofmt := gofmt
else
  gofmt := cat `\# (cat - poor-man's gofmt)`
endif


# ---------------- some implementation details ---------------- #

.NOTINTERMEDIATE:
.DELETE_ON_ERROR:

self := $(lastword $(MAKEFILE_LIST))
  # this makefile's filename

$(self):: ;
  # don't rebuild the makefile itself


# ------------------------ repo-restore ----------------------- #

dir-entries-actual  = $(filter-out . .. .vscode,$(wildcard * .* */*))
dir-entries-want    = $(self) docsgen docsgen/docsgen.go

dir-entries-missing = $(filter-out $(dir-entries-actual),$(dir-entries-want))
dir-entries-extra		= $(filter-out $(dir-entries-want)  ,$(dir-entries-actual))

repo-restore: clean
	@echo $@: checking cleaned state...
	@[ -z '$(dir-entries-missing)' ] || { echo 'FAIL: files missing'; exit 1; }
	@[ -z '$(dir-entries-extra)'   ] || { echo 'FAIL: extra files'  ; exit 1; }
	@echo ...cleaned state OK. Files:
	tree -F --noreport
	@echo
	@echo $@: building...
	$(MAKE) all >/dev/null
	@echo
	@echo ...building DONE.; echo
	$(MAKE) clean-aux
	@echo


# ------------------------- self-tests ------------------------ #

all-tests := \
  self-test_makeall       \
	self-test_shuffle       \
  self-test_no-rerun-init \
  self-test_tidy          \
	self-lint

.PHONY:\
self-test
self-test:
	@: \
	$(foreach test,$(all-tests),\
	  $(shell \
	    >&2 printf '%-30s' $(test); \
	    $(MAKE) $(test) >/dev/null && >&2 echo OK || { >&2 echo FAIL; exit 1; }; \
	   ) \
	 )
	$(MAKE) clean all >/dev/null
    # after tests, leave dir with sources created 

define assert_oneCobraInit
  ggrep -c -P 'cobra-cli\b.*\binit' \
  | { read n; \
      [ $$n != 1 ] && { echo FAIL: $$n cobra init, expected 1; exit 1; }; \
      echo OK; }
endef

define assert_binsExecution
  @ ./$(main-bin)    --help >/dev/null
	@ ./$(docsgen-bin) --help >/dev/null
  @ [ "$$(./$(main-bin) boodbye)" = 'boodbye called' ]
endef

define check_gofumpt
	if ! command -v gofumpt >/dev/null; then \
	  echo Warning: no gofumpt; \
	else \
	  diff=$$(gofumpt -d .); \
	  if [ -n "$$diff" ]; then \
	    echo Warning: non-empty gofumpt diff; \
	    printf 'first line:\n"%s"\n' "$$(<<< "$$diff" head -n1)"; \
	  else \
	    echo OK: no gofumpt diff; \
	    typeset -p diff; \
	  fi; \
	fi # >&2
endef

can-shuffle := $(shell $(MAKE) --help | ggrep -q -- '--shuffle' && echo 1)

.PHONY:\
self-test_makeall
self-test_makeall:
	$(MAKE) clean all      >/dev/null
	$(check_gofumpt)
	$(MAKE) clean binaries >/dev/null
	$(assert_binsExecution)
	$(MAKE) clean
	@echo 'OK'
	@echo

.PHONY:\
self-test_shuffle
ifeq '$(can-shuffle)' '1'
self-test_shuffle:
	$(MAKE) --shuffle=1 clean all >/dev/null
	$(MAKE) --shuffle=2 clean all >/dev/null
	$(MAKE) --shuffle=3 clean all >/dev/null
	$(MAKE) --shuffle=4 clean all >/dev/null
	@echo 'OK'
	@echo
else
self-test_shuffle:
	@printf '(`make --shuffle` not supported, skipping) ' >&2
	@echo
endif

.PHONY:\
self-test_no-rerun-init
self-test_no-rerun-init:
	# one make invocation:
	$(MAKE) clean main.go cmd/hello.go \
	| $(assert_oneCobraInit)
	$(MAKE) clean
	# separate make invocations:
	{ $(MAKE) clean main.go; $(MAKE) cmd/hello.go; } \
	| $(assert_oneCobraInit)
	# done:
	$(MAKE) clean

.PHONY:\
self-test_tidy
self-test_tidy:
	$(MAKE) clean binaries
	rm go.sum cmd/hello.go
	$(MAKE) $(main-bin) && echo 'OK' || { echo 'FAILED'; exit 1; }
	$(MAKE) clean
	@echo

.PHONY:\
self-lint
self-lint: grep-tt-lines = ggrep -P --line-number $1 '^\t\t' $(self)
self-lint:
	@if $(call grep-tt-lines,--silent); then \
	    echo $@: FAILED, lines:;   \
	    $(call grep-tt-lines,) | cat -t; \
	    exit 1; \
	fi
	@echo 'OK'
	@echo

