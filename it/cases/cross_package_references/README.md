# Cross Package References Test Case

## Purpose

This test case specifically verifies that cross-package references work correctly with package renaming. It ensures that when packages are renamed via annotations, all references to those packages are also updated, not just the package declaration itself.

## Test Scenario

### Input Files

1. **`types/v1/types.proto`** (package `types.v1` → `public_types.v1`)
   - Defines `User` message with a private field
   - Defines `Status` enum with a private value
   - Uses annotation: `option (cleanapi.file).package = "public_types.v1";`

2. **`services/v1/services.proto`** (package `services.v1` → `public_services.v1`)
   - Imports and references types from `types.v1` package
   - Uses `types.v1.User` and `types.v1.Status` in various field types:
     - Regular fields: `types.v1.User user = 1;`
     - Repeated fields: `repeated types.v1.User related_users = 3;`
     - Optional fields: `optional types.v1.User requesting_user = 4;`
     - Map fields (value): `map<string, types.v1.User> user_map = 4;`
     - Map fields (key): `map<int32, types.v1.Status> status_map = 5;`
   - Has private elements (field and method)
   - Uses annotation: `option (cleanapi.file).package = "public_services.v1";`

### Expected Behavior

The generator should:

1. **Rename package declarations**: 
   - `types.v1` → `public_types.v1`
   - `services.v1` → `public_services.v1`

2. **Update import paths**:
   - `import "types/v1/types.proto"` → `import "public_types/v1/types.proto"`

3. **Update type references in all field contexts**:
   - Regular: `types.v1.User` → `public_types.v1.User`
   - Repeated: `repeated types.v1.User` → `repeated public_types.v1.User`
   - Optional: `optional types.v1.User` → `optional public_types.v1.User`
   - Map values: `map<string, types.v1.User>` → `map<string, public_types.v1.User>`
   - Map keys: `map<int32, types.v1.Status>` → `map<int32, public_types.v1.Status>`

4. **Transform file paths**:
   - `types/v1/types.proto` → `public_types/v1/types.proto`
   - `services/v1/services.proto` → `public_services/v1/services.proto`

5. **Remove private elements**: Private fields, enum values, and methods should be filtered out

6. **Remove annotations**: Import of `annotations.proto` and file-level options should be removed

## Key Verification Points

- ✅ Package names are renamed in both files
- ✅ Import statement is updated to reference the new package path
- ✅ All type references (`types.v1.*`) are updated to use the new package name (`public_types.v1.*`)
- ✅ Output file paths match the renamed package structure
- ✅ Private elements are correctly filtered
- ✅ Annotation imports and options are removed from output

