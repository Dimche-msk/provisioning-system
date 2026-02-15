# Version Management Skill

## Description
This skill handles the unified versioning of the Provisioning System. The project uses a single source of truth for versions to ensure consistency between the backend, frontend, and build artifacts.

## Single Source of Truth
The version is stored in the `VERSION` file at the project root.

## How to Update Version
1.  **Modify the `VERSION` file**: Open the file at the project root (`/VERSION`) and update the version string (e.g., `1.1.1`).
2.  **Run the update script**: From the project root, execute the following command:
    ```bash
    ./scripts/update-version.sh
    ```
    This script automatically synchronizes the version to:
    -   `backend/internal/version/version.go`
    -   `frontend/src/lib/version.ts`
    -   `frontend/package.json`

## Rules
-   **NEVER** update versions manually in individual files.
-   **ALWAYS** use the `scripts/update-version.sh` script after modifying the `VERSION` file.
-   Ensure both backend and frontend are buildable after a version sync.

# Development Environment Credentials

The following credentials should be used by the agent and browser subagents to access the dev system interface:

- **Login**: `admin`
- **Password**: `password123`

> [!IMPORTANT]
> Use these credentials whenever a login screen is encountered during browser verification tasks.