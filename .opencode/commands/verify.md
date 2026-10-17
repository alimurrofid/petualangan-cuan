---
description: Run the project verification loop for the areas touched by the current changes
---

Load the `verify` skill and follow its procedure exactly: detect which areas were touched (`git status --short`), run each touched area's verification commands, fix real failures, and report ✅/❌ per component.
Never report done before the touched areas are green.
