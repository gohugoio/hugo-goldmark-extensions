#!/usr/bin/env python3
"""
Extract test case markdown from passthrough_test.go for manual testing.

This script extracts the 'input' variable from each test function and generates
a markdown file containing all test cases. Each test case is shown twice:
once in a code block and once rendered, for easy manual verification in
markdown previewers.

Usage:
    python3 extract_test_cases.py [input_file] [output_file]

Default:
    python3 extract_test_cases.py
    # Reads: passthrough/passthrough_test.go
    # Writes: passthrough/test-cases.md
"""

import re
import sys
from pathlib import Path


def extract_test_cases(input_file):
    """Extract test cases from Go test file."""
    with open(input_file, 'r', encoding='utf-8') as f:
        content = f.read()

    # Pattern to match test functions and their input variables
    # Captures: test name, delimiter type, and content
    # Group 1: Test name, Group 2: delimiter (` or "), Group 3: content
    pattern = r'func (Test\w+)\(t \*testing\.T\) \{.*?input := ([`"])([^`"]*)\2'
    matches = re.findall(pattern, content, re.DOTALL)

    return matches


def generate_markdown(test_cases, output_file):
    """Generate markdown file with all test cases."""
    with open(output_file, 'w', encoding='utf-8') as f:
        f.write("# Passthrough Extension Test Cases\n\n")
        f.write("This file contains all test inputs from the passthrough test suite for manual testing.\n\n")

        for test_name, delimiter, markdown_input in test_cases:
            # Only unescape interpreted string literals (double-quoted)
            # Raw string literals (backticked) should preserve backslashes verbatim
            if delimiter == '"':
                markdown_input = markdown_input.replace('\\n', '\n').replace('\\"', '"').replace('\\\\', '\\')

            # Write the test section
            f.write(f"## {test_name}\n\n")
            f.write("```markdown\n")
            f.write(markdown_input)
            f.write("\n```\n\n")
            f.write(markdown_input)
            f.write("\n\n")

    return len(test_cases)


def main():
    # Parse command line arguments with defaults relative to script location
    script_dir = Path(__file__).parent
    passthrough_dir = script_dir.parent
    default_input = passthrough_dir / 'passthrough_test.go'
    default_output = passthrough_dir / 'test-cases.md'

    input_file = sys.argv[1] if len(sys.argv) > 1 else str(default_input)
    output_file = sys.argv[2] if len(sys.argv) > 2 else str(default_output)

    # Verify input file exists
    if not Path(input_file).exists():
        print(f"Error: Input file '{input_file}' not found", file=sys.stderr)
        sys.exit(1)

    # Extract test cases
    print(f"Extracting test cases from {input_file}...", file=sys.stderr)
    test_cases = extract_test_cases(input_file)

    if not test_cases:
        print("Warning: No test cases found", file=sys.stderr)
        sys.exit(1)

    # Generate output
    count = generate_markdown(test_cases, output_file)
    print(f"Extracted {count} test cases to {output_file}", file=sys.stderr)

    # Report file size
    size = Path(output_file).stat().st_size
    print(f"Output file size: {size:,} bytes", file=sys.stderr)


if __name__ == '__main__':
    main()
