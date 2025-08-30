#!/bin/bash
set -euo pipefail

# Get the latest annotated tag
TAG=$(git describe --tags --abbrev=0)

echo "Creating release for tag: $TAG"

# Push commits and the tag
git push origin main
git push origin "$TAG"

# Extract release notes from CHANGELOG.md for this tag
# (assumes CHANGELOG sections start with '## vX.Y.Z')
awk "/^## $TAG\$/,/^## v/" CHANGELOG.md | sed '$d' > RELEASE_NOTES.md || true

# If this was the last section (no following ## v...), awk will grab until EOF
if [ ! -s RELEASE_NOTES.md ]; then
  awk "/^## $TAG\$/,/^\$/ {print}" CHANGELOG.md > RELEASE_NOTES.md
fi

echo "Release notes for $TAG written to RELEASE_NOTES.md"
echo "Now you can create a GitHub release using this file."
