#!/bin/sh
fail_if_err() {
	eval $2
	if [ "$?" -ne 0 ]; then
		echo "FAILED " $1
		exit 1
	fi
}

fail_if_err "FORMAT" "[ -z $(goimports -l .) ]"
fail_if_err "TEST" "go test ./... > /dev/null"
fail_if_err "VET" "go vet ./..."
fail_if_err "LINT" "golint -set_exit_status \$(go list ./...)"
fail_if_err "CYCLO" "gocyclo -over 30 ."

# static check
# honnef.co/go/tools/cmd/staticcheck
# complexity check
# github.com/fzipp/gocyclo/cmd/gocyclo
# from Fyne Conf 2022 https://youtu.be/J8960TmU2jY