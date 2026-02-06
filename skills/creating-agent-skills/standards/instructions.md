# Writing Effective Instructions

## Purpose

Guidelines for writing skill instructions that are concise, clear, and appropriate for agent capabilities.

## Core Principle: Concise is Key

The context window is a shared resource. Only add context the agent doesn't already have.

Challenge each piece of information:
- "Does the agent really need this explanation?"
- "Can I assume the agent knows this?"
- "Does this paragraph justify its token cost?"

### Good Example (~50 tokens)

```markdown
## Extract PDF text

Use pdfplumber for text extraction:

```python
import pdfplumber

with pdfplumber.open("file.pdf") as pdf:
    text = pdf.pages[0].extract_text()
```
```

### Bad Example (~150 tokens)

```markdown
## Extract PDF text

PDF (Portable Document Format) files are a common file format that contains
text, images, and other content. To extract text from a PDF, you'll need to
use a library. There are many libraries available for PDF processing, but we
recommend pdfplumber because it's easy to use and handles most cases well.
First, you'll need to install it using pip. Then you can use the code below...
```

The concise version assumes the agent knows what PDFs are and how libraries work.

## Degrees of Freedom

Match specificity to task fragility and variability.

### High Freedom (Text-Based Instructions)

Use when:
- Multiple approaches are valid
- Decisions depend on context
- Heuristics guide the approach

```markdown
## Code review process

1. Analyze the code structure and organization
2. Check for potential bugs or edge cases
3. Suggest improvements for readability and maintainability
4. Verify adherence to project conventions
```

### Medium Freedom (Pseudocode or Parameters)

Use when:
- A preferred pattern exists
- Some variation is acceptable
- Configuration affects behavior

```markdown
## Generate report

Use this template and customize as needed:

```python
def generate_report(data, format="markdown", include_charts=True):
    # Process data
    # Generate output in specified format
    # Optionally include visualizations
```
```

### Low Freedom (Specific Scripts)

Use when:
- Operations are fragile and error-prone
- Consistency is critical
- A specific sequence must be followed

```markdown
## Database migration

Run exactly this script:

```bash
python scripts/migrate.py --verify --backup
```

Do not modify the command or add additional flags.
```

## Avoid Common Anti-Patterns

### Don't Offer Too Many Options

**Bad:**
```markdown
You can use pypdf, or pdfplumber, or PyMuPDF, or pdf2image, or...
```

**Good:**
```markdown
Use pdfplumber for text extraction.

For scanned PDFs requiring OCR, use pdf2image with pytesseract instead.
```

### Don't Include Time-Sensitive Information

**Bad:**
```markdown
If you're doing this before August 2025, use the old API.
After August 2025, use the new API.
```

**Good:**
```markdown
## Current method

Use the v2 API endpoint: `api.example.com/v2/messages`

## Old patterns

<details>
<summary>Legacy v1 API (deprecated)</summary>
The v1 API used: `api.example.com/v1/messages`
</details>
```

### Use Consistent Terminology

**Bad:**
- Mix "API endpoint", "URL", "API route", "path"
- Mix "field", "box", "element", "control"
- Mix "extract", "pull", "get", "retrieve"

**Good:**
- Always "API endpoint"
- Always "field"
- Always "extract"

## Workflows

For complex tasks, break into clear sequential steps with checklists:

```markdown
## Research synthesis workflow

Copy this checklist and track progress:

```
Research Progress:
- [ ] Step 1: Read all source documents
- [ ] Step 2: Identify key themes
- [ ] Step 3: Cross-reference claims
- [ ] Step 4: Create structured summary
- [ ] Step 5: Verify citations
```

**Step 1: Read all source documents**
Review each document. Note main arguments and evidence.

**Step 2: Identify key themes**
Look for patterns. What themes appear repeatedly?
...
```

## Templates

Provide templates for consistent output:

```markdown
## Report structure

Use this template:

```markdown
# [Analysis Title]

## Executive summary
[One-paragraph overview]

## Key findings
- Finding 1 with supporting data
- Finding 2 with supporting data

## Recommendations
1. Specific actionable recommendation
2. Specific actionable recommendation
```
```

## Examples

For quality-dependent outputs, provide input/output pairs:

```markdown
## Commit message format

**Example 1:**
Input: Added user authentication with JWT tokens
Output:
```
feat(auth): implement JWT-based authentication

Add login endpoint and token validation middleware
```

**Example 2:**
Input: Fixed bug where dates displayed incorrectly
Output:
```
fix(reports): correct date formatting in timezone conversion

Use UTC timestamps consistently across report generation
```
```

## Stop Points

Mark where user input is required:

```markdown
### Step 2: Gather Requirements 🛑

Collect from user:
1. Primary language and version
2. Frameworks in use
3. Additional tools needed

**🛑 STOP**: Wait for user to provide requirements.
```
