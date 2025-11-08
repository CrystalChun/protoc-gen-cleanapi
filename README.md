# Public API generator

This project includes a custom protobuf code generator that creates a filtered
_public_ version of the API by removing elements marked as private. This allows
you to maintain internal/private fields, methods, and services in your proto
definitions while generating a clean public API.

## Quick start

1. **Build the plugin:**

   ```bash
   go build
   ```

2. **Mark elements as private in your proto files:**

   ```protobuf
   import "cleanapi/annotations.proto";

   message Host {
     string id = 1;

     // This field will not appear in the public API
     string internal_status = 2 [(cleanapi.field).private = true];
   }
   ```

3. **Generate the clean API:**

   ```bash
   protoc \
   --proto_path=proto \
   --proto_path=<your-proto-dir> \
   --plugin=protoc-gen-cleanapi=./protoc-gen-cleanapi \
   --public_out=<output-dir> \
   --public_opt=proto_root=<your-proto-dir> \
   <your-proto-dir>/*.proto
   ```

   This creates filtered proto files with:
   - All private elements removed
   - Formatting and comments preserved
   - Package names transformed (configurable in the generator)

## Examples

For complete usage examples, see the integration tests in the `it` directory. Each sub-directory of
the `cases` directory is a test case and
contains the following sub-directories:

- `private` - Example proto files with private annotations.
- `public` - Expected output after filtering.

## Configuration

The plugin can be configured using environment variables with the `PROTOC_GEN_PUBLIC_API_` prefix:

- **`PROTOC_GEN_PUBLIC_API_LOG_FILE`**: Path to the log file. Special values:
  - `stderr`: Write logs to standard error
  - `stdout`: Write logs to standard output (only for debugging, interferes with protoc)
  - Default: `/tmp/log.txt`
  - Any other value: Path to a file where logs will be written
  
- **`PROTOC_GEN_PUBLIC_API_LOG_LEVEL`**: Logging level. Valid values:
  - `debug`: Most verbose logging
  - `info` (default): Informational messages
  - `warn`: Warning messages only
  - `error`: Error messages only

- **`PROTOC_GEN_PUBLIC_API_DUMP_REQUEST`**: Path where the `CodeGeneratorRequest` will be dumped for debugging

- **`PROTOC_GEN_PUBLIC_API_DUMP_RESPONSE`**: Path where the `CodeGeneratorResponse` will be dumped for debugging

Example:

```bash
PROTOC_GEN_PUBLIC_API_LOG_FILE=/tmp/plugin.log \
PROTOC_GEN_PUBLIC_API_LOG_LEVEL=debug \
protoc --plugin=protoc-gen-cleanapi=./protoc-gen-cleanapi ...
```

## Debugging

For debugging purposes, you can dump the `CodeGeneratorRequest` to a file and then replay it:

1. **Capture a request during protoc execution:**

   Set the `PROTOC_GEN_PUBLIC_API_DUMP_REQUEST` environment variable to the path where you want to save the request:

   ```bash
   $ PROTOC_GEN_PUBLIC_API_DUMP_REQUEST=/tmp/request.bin protoc \
   --proto_path=proto \
   --proto_path=it/cases/basic_filtering/input \
   --plugin=protoc-gen-cleanapi=./protoc-gen-cleanapi \
   --public_out=output \
   --public_opt=proto_root=it/cases/basic_filtering/input \
   example.proto
   ```

   This saves the serialized `CodeGeneratorRequest` to `/tmp/request.bin`.
   
   Optionally, you can also capture the response by setting `PROTOC_GEN_PUBLIC_API_DUMP_RESPONSE`:

   ```bash
   $ PROTOC_GEN_PUBLIC_API_DUMP_REQUEST=/tmp/request.bin \
   PROTOC_GEN_PUBLIC_API_DUMP_RESPONSE=/tmp/response.bin \
   protoc \
   --proto_path=proto \
   --proto_path=it/cases/basic_filtering/input \
   --plugin=protoc-gen-cleanapi=./protoc-gen-cleanapi \
   --public_out=output \
   --public_opt=proto_root=it/cases/basic_filtering/input \
   example.proto
   ```
   
   Note: You need both proto paths because `example.proto` imports `cleanapi/annotations.proto`
   from the `proto/` directory.

2. **Replay the request directly for debugging:**

   ```bash
   ./protoc-gen-cleanapi < /tmp/request.bin > /tmp/response.bin
   ```

   This allows you to run the plugin directly without invoking protoc, which is useful for debugging
   with tools like `dlv` or adding debug logging.

## Documentation

For detailed information about the available annotations and how to use them, see the
[annotations.proto](proto/cleanapi/annotations.proto) file.
