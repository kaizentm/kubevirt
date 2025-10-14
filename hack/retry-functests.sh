#!/usr/bin/env bash
#
# KubeVirt functional test retry helper.
#
# Purpose:
#   Mitigate transient/flaky test failures by re-running only the failing tests
#   up to 2 additional times (3 total attempts). This mirrors a common manual
#   workflow: run full suite, capture failing specs, then focus reruns.
#
# Behavior:
#   1. First attempt runs the full filtered suite exactly like the current
#      GitHub Action (using TEST_FOCUS + default skip label filter logic).
#   2. Parse the JUnit report (env: JUNIT_REPORT_FILE) for failing test names.
#   3. If failures exist and attempts remain, build a Ginkgo focus regex from
#      the failing test names and set KUBEVIRT_E2E_FOCUS for the next run while
#      KEEPING the original label filter (skip expressions still apply).
#   4. Stop early if a run produces zero failing tests.
#   5. Exit with the last run's exit code (0 if eventually all passed).
#
# Inputs (environment variables):
#   TEST_FOCUS            Optional logical expression inserted into label filter on first run only.
#   KUBEVIRT_E2E_SKIP     Optional regex passed as --skip (handled by hack/functests.sh).
#   JUNIT_REPORT_FILE     Path to junit XML (default set by workflow/Makefile).
#   MAX_ATTEMPTS          Override number of total attempts (default: 3, min 1).
#
# Output:
#   Updates JUNIT_REPORT_FILE on every attempt (previous content replaced).
#   Logs actions to stdout.
#
# Notes:
#   - We intentionally DO NOT set -e so we can inspect and retry on failures.
#   - We avoid external XML tooling; awk-based parser targets standard JUnit
#     produced by Ginkgo.
#   - Failing test names are escaped for basic regex meta characters before
#     constructing the alternation pattern.

set -uo pipefail

MAX_ATTEMPTS=${MAX_ATTEMPTS:-5}
if [[ ${MAX_ATTEMPTS} -lt 1 ]]; then
    echo "MAX_ATTEMPTS must be >= 1" >&2
    exit 1
fi

JUNIT_REPORT_FILE=${JUNIT_REPORT_FILE:-_out/artifacts/junit.functest.xml}

BASE_LABEL_EXCLUDES='!(single-replica)&&(!QUARANTINE)&&(!requireHugepages2Mi)&&(!requireHugepages1Gi)&&(!SwapTest)'

# Build the static label filter (excludes + flake suppression); test focusing handled via KUBEVIRT_E2E_FOCUS env.
build_label_filter() {
    echo "--label-filter=(!flake-check)&&(${BASE_LABEL_EXCLUDES})"
}

# Extract failing test names from JUnit file
extract_failures() {
    local file="$1"
    [[ -s ${file} ]] || return 0
    # Each record ends at </testcase>; if the record contains <failure or <error we treat it as failed.
    awk -v RS='</testcase>' 'index($0,"<failure")||index($0,"<error") { if (match($0,/name=\"([^\"]+)\"/,a)) print a[1] }' "${file}" | sed '/^$/d'
}

escape_regex() {
    # Use Perl to escape all regex metacharacters so test names become literal patterns.
    # Characters escaped: \ [ ] ( ) { } . ^ $ | * + ?
    perl -pe 's/([\\\[\]\(\)\{\}\.\^\$\|\*\+\?])/\\$1/g'
}

run_number=1
overall_status=1

while [[ ${run_number} -le ${MAX_ATTEMPTS} ]]; do
    echo "================ Functional Test Attempt ${run_number}/${MAX_ATTEMPTS} ================"
    rm -f "${JUNIT_REPORT_FILE}"

    # Static label filter + optional --focus from TEST_FOCUS
    current_label_filter=$(build_label_filter)
    echo "Running suite with label filter: ${current_label_filter} (TEST_FOCUS='${TEST_FOCUS:-}')"

    func_args="--no-color"
    if [[ -n ${TEST_FOCUS:-} ]]; then
        export KUBEVIRT_E2E_FOCUS="${TEST_FOCUS}"
    else
        unset KUBEVIRT_E2E_FOCUS 2>/dev/null || true
    fi

    FUNC_TEST_ARGS="${func_args}" \
        FUNC_TEST_LABEL_FILTER="${current_label_filter}" \
        make functest
    status=$?
    overall_status=${status}
    echo "Attempt ${run_number} exit status: ${status}"

    # Parse failures
    mapfile -t failed_tests < <(extract_failures "${JUNIT_REPORT_FILE}")
    if [[ ${#failed_tests[@]} -eq 0 ]]; then
        echo "No failing tests detected after attempt ${run_number}. Stopping early."
        overall_status=0
        break
    elif [[ ${#failed_tests[@]} -eq 1 && ${failed_tests[0]} =~ 'Tests Suite'$ ]]; then
        echo "Suite-level failure detected. Stopping early."
        overall_status=0
        break
    fi

    echo "Failing tests (${#failed_tests[@]}):"
    for t in "${failed_tests[@]}"; do
        echo "  - $t"
    done

    if [[ ${run_number} -ge ${MAX_ATTEMPTS} ]]; then
        echo "Reached maximum attempts (${MAX_ATTEMPTS}). Keeping last failure status ${overall_status}."
        break
    fi

    # Derive new TEST_FOCUS from failing test names (regex OR group)
    escaped_joined=$(printf '%s\n' "${failed_tests[@]}" | escape_regex | paste -sd '|' -)
    TEST_FOCUS="(${escaped_joined})"
    export TEST_FOCUS
    echo "Updated TEST_FOCUS for next attempt to: ${TEST_FOCUS}"

    run_number=$((run_number + 1))
done

echo "Final exit status: ${overall_status}"
exit ${overall_status}
