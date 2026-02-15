#!/bin/bash
# scripts/update-version.sh

# Get absolute path of the script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$( dirname "$SCRIPT_DIR" )"

cd "$PROJECT_ROOT"

if [ ! -f VERSION ]; then
    echo "Error: VERSION file not found in $PROJECT_ROOT"
    exit 1
fi

# Get version from VERSION file
VERSION=$(cat VERSION | tr -d '[:space:]')

echo "Updating project version to $VERSION..."

# Update Backend
cat <<EOF > backend/internal/version/version.go
package version

var (
	Version = "$VERSION"
)
EOF
echo "Backend version updated in backend/internal/version/version.go"

# Update Frontend (Typescript constant)
cat <<EOF > frontend/src/lib/version.ts
export const APP_VERSION = "$VERSION";
EOF
echo "Frontend version updated in frontend/src/lib/version.ts"

# Update frontend/package.json
# Use -i '' for macOS sed compatibility
sed -i '' "s/\"version\": \".*\"/\"version\": \"$VERSION\"/" frontend/package.json
echo "Frontend package.json updated"

echo "Version synchronization complete."
