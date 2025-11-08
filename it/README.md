# Integration tests

This directory contains integration tests for the `protoc-gen-cleanapi` plugin. These are end-to-end tests
that verify the plugin works correctly when integrated with `protoc`, testing the complete workflow from input
proto files through the protoc plugin system to generated output.

## What are integration tests?

Unlike unit tests that test individual components in isolation, these integration tests:

- Execute the actual `protoc` command with the plugin.
- Process real proto files through the complete toolchain.
- Verify the final output matches expectations.
- Test the plugin in the same way users will invoke it.

This ensures the plugin works correctly as part of the protoc ecosystem, not just as standalone code.

## Structure

Test cases are located in the `cases/` subdirectory. Each test case directory must contain:

- `input/` - Input proto files containing a mix of public and private content.
- `output/` - Expected output proto files after filtering with the plugin.

## Running the tests

The tests are written using the _Ginkgo_ framework and located in `it_test.go`.

To run the integration tests:

```bash
$ cd it
$ ginkgo run
```

## Adding new test cases

To add a new test case:

1. Create a new directory under `it/cases` with a descriptive name.

2. Create an `input` subdirectory with input proto files.

3. Create an `output` subdirectory with the expected filtered output.

4. Add a _Entry_ to the `it_test.go` with a human friendly name and the name of the new directory.

5. Run the tests to verify.

## How it works

For each test case the `it_test.go` file will do the following:

- Run `protoc` with the plugin on all `.proto` files in `input`.

- Compare the generated output with the expected files in `output`.

- Verify that all expected files are generated.

- Verify that no extra files are generated.

## Notes

- The tests require `protoc` to be installed and available in the `PATH`.

