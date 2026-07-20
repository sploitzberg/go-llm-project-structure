#!/usr/bin/env bash
# scripts/ci/pre-commit/17-struct-fields.sh
# Validate that domain structs don't have external dependencies in fields

set -euo pipefail

echo "> Running struct field validation"

if [ ! -d "internal/core/domain" ]; then
    echo "No domain directory found, skipping"
    exit 0
fi

# Find all Go files in domain
domain_files=$(find internal/core/domain -name "*.go" 2>/dev/null || true)

errors=0

for file in $domain_files; do
    # Find struct definitions
    struct_defs=$(grep -n "^type [A-Z].* struct" "$file" || true)

    while IFS= read -r line; do
        [[ -n "$line" ]] || continue
        line_num=$(echo "$line" | cut -d: -f1)
        struct_line=$(echo "$line" | cut -d: -f2-)
        struct_name=$(echo "$struct_line" | awk '{print $2}')

        # Extract struct body
        struct_body=$(sed -n "${line_num},/^}/p" "$file" || true)

        # Check for field types that might be external dependencies
        # Look for types from other internal packages
        if echo "$struct_body" | grep -E "internal/(adapter|core/services|core/ports)" | grep -q .; then
            echo "error: $file:$line_num Struct $struct_name has fields from adapter/services/ports packages"
            echo "$struct_body" | grep -E "internal/(adapter|core/services|core/ports)"
            errors=$((errors + 1))
        fi

        # Check for framework types in domain structs
        if echo "$struct_body" | grep -E "(github\.com/gin|github\.com/gorilla|database/sql|http\.ResponseWriter|http\.Request)" | grep -q .; then
            echo "error: $file:$line_num Struct $struct_name has framework types (forbidden in domain)"
            echo "$struct_body" | grep -E "(github\.com/gin|github\.com/gorilla|database/sql|http\.ResponseWriter|http\.Request)"
            errors=$((errors + 1))
        fi
    done <<< "$struct_defs"
done

# Also check service and port for framework leaks
for dir in internal/core/services internal/core/ports; do
    if [ ! -d "$dir" ]; then
        continue
    fi

    files=$(find "$dir" -name "*.go" 2>/dev/null || true)

    for file in $files; do
        struct_defs=$(grep -n "^type [A-Z].* struct" "$file" || true)

        while IFS= read -r line; do
            [[ -n "$line" ]] || continue
            line_num=$(echo "$line" | cut -d: -f1)
            struct_body=$(sed -n "${line_num},/^}/p" "$file" || true)

            # Check for framework types in services/ports
            if echo "$struct_body" | grep -E "(github\.com/gin|github\.com/gorilla|database/sql|http\.ResponseWriter|http\.Request)" | grep -q .; then
                echo "error: $file:$line_num Struct in $dir has framework types (forbidden)"
                errors=$((errors + 1))
                echo "$struct_body" | grep -E "(github\.com/gin|github\.com/gorilla|database/sql|http\.ResponseWriter|http\.Request)"
            fi
        done <<< "$struct_defs"
    done
done

if ((errors > 0)); then
    echo "error: Struct field validation failed with $errors error(s)"
    exit 1
fi

echo "Struct fields: OK"
echo
