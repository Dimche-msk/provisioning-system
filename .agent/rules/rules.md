# Agent Rules

## Verification Procedures
If it is necessary to check the operation of a feature or fix:
1. **Ask the user** for assistance for ANY UI or visual verification.
2. **Do NOT use autonomous browser subagents** to verify visual changes.
3. **Explain clearly** what test needs to be performed.
4. **Specify** what logs or screenshots should be provided for verification.

> [!IMPORTANT]
> Never use the browser_subagent for visual UI verification. Always ask the user to verify UI changes manually on the development server (port 5454) or the production static build (port 8080).

## Task Focus and UI Consistency
1. **Stick to the task**: Do not modify parts of the code that are not directly related to the current objective.
2. **Preserve UI defaults**: Do not change placeholders, labels, or default selection values (e.g., "Все" for filters) unless explicitly requested.
3. **Be conservative**: When refactoring, maintain existing UI patterns and localized strings.
16: 
17: ## Version Management
1. **Unified Versioning**: Follow the procedures in the `Version Management` skill to update project versions.
