#!/bin/bash
set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 1.2.0"
    exit 1
fi

VERSION=$1

echo "Bumping version to $VERSION..."

# Update version in main.go or version file if exists
# Update CHANGELOG.md
# Update README.md badges if needed
# Create git tag
git add -A
git commit -m "chore: bump version to $VERSION"
git tag -a "v$VERSION" -m "Release v$VERSION"

echo "✅ Version bumped to $VERSION"
echo "To release: git push && git push origin v$VERSION"
