---
description: Rule for automatic versioning
---

# Versioning Rule

Every time a code change is made to the backend or frontend, the version number MUST be incremented.

## Frontend Versioning
- **Location**: `frontend/src/lib/components/Sidebar.svelte`
- **Pattern**: `<p class="text-xs text-center text-gray-500">vX.Y.Z</p>`
- **Action**: Increment the patch version (Z) for minor changes, or Y for major feature changes.

## Backend Versioning
- **Location**: `backend/internal/version/version.go`
- **Pattern**: `Version = "X.Y.Z"`
- **Action**: Increment the patch version (Z) for minor changes, or Y for major feature changes.

## Communication
- Always inform the user about the new versions in the final notification.
