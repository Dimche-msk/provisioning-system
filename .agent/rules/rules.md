# Agent Rules

## Verification Procedures
If it is necessary to check the operation of a feature or fix:
1. **Ask the user** for assistance.
2. **Explain clearly** what test needs to be performed.
3. **Specify** what logs or screenshots should be provided for verification.

> [!IMPORTANT]
> Do not attempt autonomous verification if it requires complex setup or credentials unless explicitly directed. When in doubt, request user verification.

## Task Focus and UI Consistency
1. **Stick to the task**: Do not modify parts of the code that are not directly related to the current objective.
2. **Preserve UI defaults**: Do not change placeholders, labels, or default selection values (e.g., "Все" for filters) unless explicitly requested.
3. **Be conservative**: When refactoring, maintain existing UI patterns and localized strings.
16: 
17: ## Version Management
1. **Unified Versioning**: Follow the procedures in the `Version Management` skill to update project versions.
