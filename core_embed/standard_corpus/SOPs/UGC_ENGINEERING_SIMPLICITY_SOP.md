# Engineering Simplicity SOP

This project is designed to be clean, fast, and easy to audit.

## Constraints
1. **Language & Idioms:** Follow the idiomatic style of the project's primary language.
2. **Dependencies:** Minimize third-party dependencies. Prefer standard libraries or lightweight, established modules.
3. **No Magic:** Do not hide control flow or error handling. Ensure errors are explicitly checked and logged.
4. **Safety & Immutability:** Do not destructively overwrite files without an explicit backup, dry-run, or user confirmation mechanism.
