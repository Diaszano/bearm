# Migrate Legacy Gemini CLI Agents to Antigravity Skills

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Convert all 38 legacy `.gemini/agents/*.md` files into Antigravity `.agents/skills/<skill-name>/SKILL.md` format.

**Architecture:** A Python helper script to parse YAML frontmatter, extract name and description, write the new clean frontmatter, and output the body to a new directory structure under `.agents/skills/`.

**Tech Stack:** Python 3, PyYAML (or basic string parsing to ensure no extra dependencies are needed).

## Global Constraints
- Do not introduce external dependencies if not present.
- Create files with exact relative paths.

---

### Task 1: Create and Run the Migration Script

**Files:**
- Create: `scratch/migrate_agents.py`
- Create: `.agents/skills/` (implicitly by script)

**Interfaces:**
- Consumes: Files in `.gemini/agents/`
- Produces: Subdirectories and `SKILL.md` files under `.agents/skills/`

- [ ] **Step 1: Write the migration python script**

Write the Python code to read and convert the frontmatter of each `.md` file from `.gemini/agents/` into `.agents/skills/<agent-name>/SKILL.md`.

```python
import os
import re

src_dir = ".gemini/agents"
dest_dir = ".agents/skills"

if not os.path.exists(dest_dir):
    os.makedirs(dest_dir)

for filename in os.listdir(src_dir):
    if not filename.endswith(".md"):
        continue
    
    src_path = os.path.join(src_dir, filename)
    with open(src_path, "r", encoding="utf-8") as f:
        content = f.read()
    
    # Parse YAML frontmatter
    # Match frontmatter between --- and ---
    match = re.match(r"^---\s*\n(.*?)\n---\s*\n(.*)$", content, re.DOTALL)
    if not match:
        print(f"Skipping {filename}: no frontmatter found")
        continue
        
    frontmatter_str, body = match.groups()
    
    # Parse name and description
    name = None
    description_lines = []
    in_description = False
    
    for line in frontmatter_str.splitlines():
        if line.startswith("name:"):
            name = line.split(":", 1)[1].strip()
            in_description = False
        elif line.startswith("description:"):
            desc_val = line.split(":", 1)[1].strip()
            description_lines.append(desc_val)
            in_description = True
        elif in_description and (line.startswith(" ") or line.startswith("\t")):
            description_lines.append(line.strip())
        else:
            in_description = False
            
    if not name:
        # Fallback to filename without extension
        name = os.path.splitext(filename)[0]
        
    description = " ".join(description_lines).strip()
    # Clean up quotes if present
    if description.startswith('"') and description.endswith('"'):
        description = description[1:-1]
    elif description.startswith("'") and description.endswith("'"):
        description = description[1:-1]
        
    # Write new skill
    skill_dir = os.path.join(dest_dir, name)
    os.makedirs(skill_dir, exist_ok=True)
    
    new_content = f"""---
name: {name}
description: >-
  {description}
---

{body.strip()}
"""
    dest_path = os.path.join(skill_dir, "SKILL.md")
    with open(dest_path, "w", encoding="utf-8") as f:
        f.write(new_content)
        
    print(f"Converted {filename} -> {dest_path}")
```

- [ ] **Step 2: Run the script to perform conversion**

Run: `python3 scratch/migrate_agents.py`
Expected: Output showing conversion of all 38 files.

- [ ] **Step 3: Verify the output files exist**

Run: `ls -la .agents/skills/`
Expected: Folders matching each agent.

- [ ] **Step 4: Remove the scratch script**

Clean up by deleting the script.

- [ ] **Step 5: Commit**

```bash
git add .agents/skills/
git commit -m "feat: migrate legacy gemini-cli agents to antigravity skills format"
```
