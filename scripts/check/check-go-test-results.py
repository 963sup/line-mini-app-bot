"""Require complete, non-skipped Go test evidence for the requested packages."""
import json
from pathlib import Path
import sys


def verify(lines, packages):
    expected = set(packages)
    finished = set()
    passed = set()
    failures = []
    for number, line in enumerate(lines, 1):
        try:
            event = json.loads(line)
        except (ValueError, TypeError) as error:
            raise ValueError(f"Invalid go test JSON at line {number}") from error
        package, test, action = event.get("Package"), event.get("Test"), event.get("Action")
        if not package or not action:
            raise ValueError(f"Missing go test event identity at line {number}")
        if test:
            if action in ("skip", "fail"):
                failures.append(f"{package}/{test}: {action}")
            if action == "pass":
                passed.add((package, test))
        elif action == "fail":
            failures.append(f"{package}: fail")
        elif action in ("pass", "skip"):
            # A package with no test files is reported as package-level skip.
            finished.add(package)
    missing = expected - finished
    if not expected or missing or failures or not passed:
        raise ValueError(f"Incomplete Go acceptance: missing={sorted(missing)}, failures={failures}, passed_tests={len(passed)}")
    return len(finished & expected), len(passed)


if __name__ == "__main__":
    try:
        with Path(sys.argv[1]).open() as report:
            packages, tests = verify(report, Path(sys.argv[2]).read_text().splitlines())
        print(f"PASS: {packages} packages completed; {tests} test/subtest results passed; no skipped test cases")
    except (ValueError, OSError) as error:
        raise SystemExit(str(error)) from error
