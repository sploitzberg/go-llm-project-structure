# Prompt Engineering Best Practices

A practical guide for writing effective prompts for LLM coding assistants. Combines guidance from Anthropic's official best practices with real-world patterns. Examples are written in the context of a Go/hexagonal architecture codebase but the principles apply universally.

---

## 1. Be Clear and Direct

The most reliable improvement you can make to any prompt is to be explicit. LLMs are like a brilliant but new employee — they lack your context, your norms, and your history. The more precisely you specify what you want, the more reliable the output.

**Golden rule:** Show your prompt to a colleague with minimal context. If they'd be confused, the model will be too.

### ❌ Vague

```text
Can you help me with the order processing logic?
```

### ✅ Specific

```text
Implement an `OrderService.Process` method in `internal/core/services/` that:
1. Validates the order is in PENDING state
2. Checks inventory availability via the `InventoryPort` interface
3. Charges the payment via the `PaymentPort` interface
4. Returns a domain error if any step fails
Do not import any HTTP or database packages — this is the services layer.
```

---

## 2. Add Context and Motivation

Explaining _why_ you want something helps the model make better decisions when instructions are ambiguous.

### ❌ No context

```text
Don't add imports from adapter packages to the domain layer.
```

### ✅ With motivation

```text
Never import adapter packages (internal/adapter/...) from the domain layer
(internal/core/domain/...). This project uses Hexagonal Architecture — the domain
must be completely isolated from external concerns. Violating this breaks the
architecture and creates circular dependencies.
```

---

## 3. Use Examples (Few-Shot / Multishot Prompting)

Examples are one of the most reliable ways to steer output format, tone, and structure. Wrap them in `<example>` or `<examples>` tags to distinguish them from instructions.

### ❌ Instruction-only

```text
Write domain errors as sentinel values.
```

### ✅ With examples

```text
Write domain errors as sentinel values following the existing pattern.

<examples>
<example>
var ErrOrderNotFound = errors.New("order not found")
var ErrInvalidOrderStatus = errors.New("invalid order status")
</example>
</examples>
```

---

## 4. Structure Prompts with XML Tags

XML tags help the model parse complex prompts unambiguously, especially when mixing instructions, context, examples, and variable inputs. Use consistent, descriptive tag names.

### ❌ Unstructured (instructions and data blended)

```text
Here is a function I wrote. It handles order processing. Review it for compliance
with our architecture rules and suggest improvements. The rules are: no adapter
imports in domain, max 4 parameters per function, no magic numbers.
```

### ✅ Structured with XML

```xml
<task>Review the following Go function for architectural compliance.</task>

<rules>
- No imports from internal/adapter/ in domain layer code
- Functions must have ≤4 parameters
- No magic numbers — use named constants
</rules>

<code>
func ProcessOrder(orderID string, userID string, itemID string, qty int, price float64) error {
    if qty > 100 {
        return errors.New("too many items")
    }
    // ...
}
</code>
```

**Common useful tags:**

| Tag                          | Purpose                      |
| ---------------------------- | ---------------------------- |
| `<instructions>`             | What to do                   |
| `<context>`                  | Background knowledge         |
| `<rules>`                    | Constraints and guardrails   |
| `<examples>` / `<example>`   | Input/output illustrations   |
| `<code>`                     | Source code to review/modify |
| `<input>`                    | Variable data for the task   |
| `<document>` / `<documents>` | Reference material           |

---

## 5. Give the Model a Role

Setting a role focuses behaviour and tone. Even one sentence makes a measurable difference.

### ❌ No role

```text
Review this code.
```

### ✅ With role

```text
You are a senior Go engineer specialising in Hexagonal Architecture and domain-driven
design. Review the following code for architectural violations, coupling issues, and
adherence to Go idioms. Be critical and direct — do not soften findings.
```

---

## 6. Tell It What To Do, Not What Not To Do

Positive instructions ("do X") are more reliable than negative instructions ("don't do Y"). When you must prohibit something, pair it with the preferred alternative.

### ❌ Negative only

```text
Do not use markdown in your response.
Do not use bullet points.
```

### ✅ Positive instruction

```text
Your response should be composed of smoothly flowing prose paragraphs. Reserve
markdown only for inline code and code blocks.
```

---

## 7. Control Format Explicitly

For structured or predictable output, specify the exact format. Use XML tags to indicate where formatted content should appear.

### ❌ Implicit format

```text
List the architecture violations you found.
```

### ✅ Explicit format

```text
List all architecture violations in the following format:

<violations>
<violation>
  <file>path/to/file.go</file>
  <line>42</line>
  <rule>No adapter imports in domain layer</rule>
  <description>Imports internal/adapter/secondary/db which is a secondary adapter</description>
</violation>
</violations>

If no violations are found, return <violations></violations>.
```

---

## 8. Long Context: Put Data First, Query Last

When working with large inputs (code reviews, multi-file analysis, documentation), structure matters.

- **Documents and code go at the top**
- **Instructions and query go at the bottom**
- This can improve response quality by up to 30% on complex inputs

### ❌ Query first (degrades quality on long inputs)

```text
Analyse these files for coupling violations. [file1...] [file2...] [file3...]
```

### ✅ Data first, query last

```xml
<documents>
<document index="1">
  <source>internal/core/domain/order.go</source>
  <document_content>
  // ... file contents
  </document_content>
</document>
<document index="2">
  <source>internal/core/services/order_service.go</source>
  <document_content>
  // ... file contents
  </document_content>
</document>
</documents>

<task>
Analyse the documents above for coupling violations. For each violation, quote the
relevant line(s) before explaining the issue.
</task>
```

---

## 9. Default-to-Action vs. Cautious Mode

Modern LLM assistants are configurable between proactive and conservative behaviour. Use the appropriate stance for your context.

### Proactive (implement by default)

```xml
<default_to_action>
By default, implement changes rather than only suggesting them. If the user's intent
is unclear, infer the most useful likely action and proceed, using tools to discover
any missing details instead of guessing.
</default_to_action>
```

### Cautious (research and recommend by default)

```xml
<do_not_act_before_instructions>
Do not jump into implementation or change files unless clearly instructed. When
intent is ambiguous, default to providing information and recommendations rather
than taking action.
</do_not_act_before_instructions>
```

**When to use each:**

| Scenario                               | Stance    |
| -------------------------------------- | --------- |
| Writing new features from a clear spec | Proactive |
| Exploring an unfamiliar codebase       | Cautious  |
| Refactoring with clear rules           | Proactive |
| Security-sensitive changes             | Cautious  |
| Automated CI/CD pipelines              | Cautious  |

---

## 10. Parallel Tool Calling

When a task involves multiple independent reads or searches, instruct the model to run them in parallel for speed.

### ❌ Sequential by default

```text
Read the order service file and then read the inventory service file.
```

### ✅ Explicit parallelism

```xml
<use_parallel_tool_calls>
If you intend to call multiple tools and there are no dependencies between them,
make all the independent calls in parallel. For example, when reading 3 files, run
3 tool calls simultaneously. Never use placeholders or guess parameters in tool calls.
</use_parallel_tool_calls>
```

---

## 11. Prevent Hallucinations — Investigate Before Answering

Models can confidently describe code they have never seen. Enforce grounding with an explicit rule.

### ❌ No grounding requirement

```text
How does the payment processing work in this codebase?
```

### ✅ With investigation requirement

```xml
<investigate_before_answering>
Never speculate about code you have not opened. If the user references a specific
file, read it before answering. Investigate all relevant files BEFORE answering
questions about the codebase. Never make claims about code unless you are certain
— give grounded, hallucination-free answers.
</investigate_before_answering>
```

---

## 12. Prevent Over-Engineering

LLMs tend to add abstractions, generalisations, and "improvements" beyond what was requested. Scope this behaviour explicitly in system prompts.

### ❌ No scope constraint

```text
Fix the bug in the ProcessOrder function.
```

### ✅ Minimal scope instruction

```xml
<minimal_changes>
Avoid over-engineering. Only make changes that are directly requested or clearly
necessary.
- Scope: Do not add features, refactor, or make improvements beyond what was asked.
- Documentation: Do not add comments or docstrings to code you did not change.
- Abstractions: Do not create helpers or utilities for one-time operations.
- Defensive coding: Do not add error handling for scenarios that cannot happen.
The right amount of complexity is the minimum needed for the current task.
</minimal_changes>
```

---

## 13. Prevent Hard-Coding and Test-Focused Solutions

Without guidance, LLMs sometimes solve for the test cases rather than the general problem.

### ❌ No constraint

```text
Make the tests pass.
```

### ✅ General solution requirement

```text
Implement a correct, general-purpose solution. Do not hard-code values or create
solutions that only work for the specific test inputs provided. Tests verify
correctness — they do not define the solution. If any test appears incorrect,
say so rather than working around it.
```

---

## 14. Safety for Irreversible Actions

For agentic systems that interact with git, databases, or external services, explicitly define what requires confirmation.

```xml
<action_safety>
Consider the reversibility and impact of your actions.
- Take local, reversible actions freely: editing files, running tests, reading code.
- Confirm before: deleting files/branches, force-pushing, dropping tables, posting to
  external services, modifying shared infrastructure.
- Never bypass safety checks (e.g. --no-verify, --force) as a shortcut.
</action_safety>
```

---

## 15. Long-Horizon Agentic Tasks

For tasks spanning multiple sessions or context windows, guide the model to persist state and not abandon work.

```xml
<long_horizon_task>
Your context window will be compacted as it approaches its limit, allowing you to
continue working indefinitely. Do not stop tasks early due to token budget concerns.
As you approach your token budget limit, save your current progress and state before
the context window refreshes. Complete tasks fully — never artificially stop early.
</long_horizon_task>
```

**Pair with a structured state file** (`progress.txt`, `tests.json`) so the model can resume reliably:

```text
At the start of each session: review progress.txt and git log.
At the end of each session: update progress.txt with completed steps and next actions.
```

---

## 16. Thinking and Reasoning Calibration

Modern models may over-think simple tasks. When you need decisive action:

### ❌ Leaves model to decide depth

```text
Implement the order validation logic.
```

### ✅ Constrained reasoning

```text
When deciding how to approach this problem, choose an approach and commit to it.
Avoid revisiting decisions unless you encounter new information that directly
contradicts your reasoning. If weighing two approaches, pick one and see it through.
```

---

## 17. The LLM-as-Judge Pattern

One of the highest-value prompt patterns: generate → review → refine. Use the model to audit its own output before you see it.

**Step 1 — Generate:**

```text
Implement the OrderService following hexagonal architecture rules.
```

**Step 2 — Self-audit (separate call or chained):**

```xml
<task>Act as a senior staff engineer performing a technical debt audit.</task>

<code>
[generated code from step 1]
</code>

<rules>
- No adapter imports in domain or services layer
- Functions ≤ 4 parameters
- No magic numbers
- All errors wrapped with %w
- No init() functions
</rules>

List every violation found. For each: file, line, rule broken, and suggested fix.
If no violations: state "No violations found."
```

**Step 3 — Fix:**

```text
Fix all violations identified in the audit. Do not change any other behaviour.
```

---

## 18. Structured Review Prompts

When reviewing code for architecture compliance, structured prompts dramatically outperform freeform ones.

### ❌ Freeform review

```text
Review this code.
```

### ✅ Checklist-driven review

```xml
<task>Review the following Go code against our architecture rules.</task>

<checklist>
1. [ ] No adapter imports in core/domain/ or core/ports/
2. [ ] Functions have ≤ 4 parameters
3. [ ] No magic numbers — named constants used
4. [ ] Errors wrapped with fmt.Errorf("...: %w", err)
5. [ ] No panic() in production code
6. [ ] No init() functions
7. [ ] context.Context is first parameter where applicable
8. [ ] Interfaces are small and focused (ideally 1-3 methods)
</checklist>

<code>
[code to review]
</code>

For each checklist item: mark PASS or FAIL. For FAILs, quote the offending line
and suggest the fix.
```

---

## Quick Reference

| Goal                       | Technique                                                   |
| -------------------------- | ----------------------------------------------------------- |
| Better accuracy            | Be explicit, add context, use examples                      |
| Structured output          | XML tags for format specification                           |
| Prevent hallucinations     | `<investigate_before_answering>`                            |
| Prevent over-engineering   | `<minimal_changes>`                                         |
| Speed up multi-file reads  | `<use_parallel_tool_calls>`                                 |
| Safety for destructive ops | `<action_safety>`                                           |
| Long-horizon tasks         | `<long_horizon_task>` + state file                          |
| Code review                | Checklist-driven structured prompt                          |
| Self-correction            | Generate → audit → fix chain                                |
| Format control             | Positive instructions, match prompt style to desired output |
