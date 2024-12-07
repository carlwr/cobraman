MAKEFLAGS     := --no-builtin-rules         \
								 --warn-undefined-variables \
								 --no-print-directory

cobra       := cobra-cli --config /dev/null
module-path := github.com/carlwr/cobraman
go-version  := 1.23


# ------------------

all-srcs :=  \
	cmd/rootCmd.go     \
	cmd/boodbye.go     \
	cmd/hello.go       \
	main.go            \
	docsgen/docsgen.go \
	| go.mod

bins :=  \
	main-bin  \
	docsgen/docsgen-bin

all     : $(all-srcs) tidy
binaries: $(bins)
.PHONY  : all binaries

.DEFAULT_GOAL := all


# ------------------

go.mod:
	> $@ \
	  $(print-gomod)
	@echo

.PHONY:\
tidy
tidy: main.go cmd/rootCmd.go | go.mod
	go mod tidy
	@echo

cmd/rootCmd.go main.go &: | go.mod
	$(cobra) init
	> cmd/rootCmd.go  \
	  cat cmd/root.go
	>>cmd/rootCmd.go  \
	  $(print-GetRootCmd)
	rm cmd/root.go
	@[ -e LICENSE ] && rm LICENSE
	@echo

cmd/hello.go   \
cmd/boodbye.go:\
cmd/%.go      : | main.go cmd/rootCmd.go
	$(cobra) add $(*F)
	@echo

$(bins): $(all-srcs) tidy

$(bins):\
%-bin  : %.go
	go build -o $@ $<
	@echo

# notes:
#
# * whenever `go.mod` is a prereq, it is only ever an _order-only prereq_
#   * i.e. `[..] : [..] | go.mod`
#   * since: go.mod is updated _in-place_ (spuriously by`cobra-cli` invocations, and when we invoke `go mod tidy`)
#
# * go.sum will be created by some of the commands, and it is sometimes updated in-place; there is no need to have `make` track this/the state of the file, hence `go.sum` is not included in the rules


# ------------------

print-gomod := \
  printf '%s\n' \
    'module $(module-path)/examples/mkbin/boodbye'  \
    'go $(go-version)'  \
    'replace $(module-path) v0.0.0 => ../../..'  \
    'require $(module-path) v0.0.0'

print-GetRootCmd := \
  printf '%s\n' \
    '// Added for use by cobraman/mkbin.'  \
    'func GetRootCmd() *cobra.Command {'   \
     $$'\treturn rootCmd'  \
    '}'



# ------------------

.PHONY:\
clean
clean: clean-bin    \
			 clean-cobra  \
			 clean-go.mod
	@echo

PHONY:\
clean-go.mod
clean-go.mod:
	rm -f go.mod go.sum

.PHONY:\
clean-cobra
clean-cobra:
	rm -rf cmd
	rm -f main.go

.PHONY:\
clean-bin
clean-bin:
	rm -rf $(bins)



# ------------------

all-tests := \
  self-test_makeall         \
	self-test_all-x2          \
  self-test_no-rerun-init-1 \
	self-test_no-rerun-init-2 \
  self-test_tidy

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
  ggrep -c 'cobra-cli .* init' \
  | { read n; \
      [ $$n != 1 ] && { echo FAIL: $$n cobra init, expected 1; exit 1; }; \
      echo OK; }
endef

define assert_binsExecution
  [ "$$(./main-bin boodbye)" = 'boodbye called' ]
  ./docsgen/docsgen-bin --help >/dev/null
endef

.PHONY:\
self-test_makeall
self-test_makeall:
	$(MAKE) clean all
	$(MAKE) clean binaries
	$(assert_binsExecution)
	$(MAKE) clean
	@echo 'OK'
	@echo

.PHONY:\
self-test_all-x2
self-test_all-x2:
	$(MAKE) clean all
	$(MAKE) all
	[   -e cmd/rootCmd.go ] || { >&2 echo FAIL: no rootCmd.go ; exit 1; }
	[ ! -e cmd/root.go    ] || { >&2 echo FAIL: root.go exists; exit 1; }
	@echo OK
	@echo

.PHONY:\
self-test_no-rerun-init-1
self-test_no-rerun-init-1:
	$(MAKE) clean cmd/boodbye.go cmd/hello.go \
	| $(assert_oneCobraInit)
	$(MAKE) clean

.PHONY:\
self-test_no-rerun-init-2
self-test_no-rerun-init-2:
	{ $(MAKE) clean cmd/boodbye.go; $(MAKE) cmd/hello.go; } \
	| $(assert_oneCobraInit)
	$(MAKE) clean

.PHONY:\
self-test_tidy
self-test_tidy:
	$(MAKE) clean binaries
	rm go.sum cmd/hello.go
	$(MAKE) main-bin && echo 'OK' || { echo 'FAILED'; exit 1; }
	$(MAKE) clean
	@echo


