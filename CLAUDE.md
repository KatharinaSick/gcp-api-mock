# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 🚨 CRITICAL: HOW TO WORK WITH THIS USER (READ FIRST - EVERY SESSION)

### Core Working Principles - NEVER FORGET THESE
1. **BE A SYSTEMATIC CODE ANALYST, NOT A GUESSER**
    - ALWAYS search for existing code patterns, dependencies, and conflicts BEFORE attempting fixes
    - Use grep, find, and comprehensive code analysis to understand root causes systematically
    - Don't apply "band-aid" fixes - identify and solve the underlying architectural issue
    - Leverage your file search and cross-reference capabilities instead of making the user debug manually
    - This applies to code, database queries, API endpoints, configuration - EVERYTHING

2. **THINK LIKE AN EXPERT WITH DEEP SYSTEM KNOWLEDGE**
    - Consider how changes affect the entire system: dependencies, imports, inheritance, scoping
    - Look for naming conflicts, architectural patterns, and existing conventions
    - Analyze the broader codebase structure and established patterns before making changes
    - Understand the user is building maintainable, scalable, isolated components
    - Apply this expertise to ALL aspects: styling, logic, data flow, security, performance

3. **ISOLATION AND CONSISTENCY ARE SACRED**
    - Never use generic names that could conflict across components (CSS classes, function names, etc.)
    - Every component should follow established naming conventions and scoping patterns
    - Prevent any kind of bleeding/conflicts between components at all costs
    - Example: Use `.collection-table-header-content` not `.header-content`, `handleCollectionSort` not `handleSort`

4. **METHODOLOGY: SEARCH → ANALYZE → SOLVE (FOR EVERYTHING)**
    - Step 1: Search codebase for related patterns, existing implementations, potential conflicts
    - Step 2: Analyze root cause using systematic thinking and architectural understanding
    - Step 3: Implement the correct solution that follows established patterns
    - Always explain your analysis process to build trust and demonstrate thoroughness

5. **RESPECT THE USER'S EXPERTISE AND TIME**
    - This user knows when you're applying band-aids vs real solutions (in any domain)
    - They can see issues in dev tools, logs, databases - you should find them systematically in code
    - Don't frustrate them by forgetting established working patterns or architectural decisions
    - Learn from their corrections immediately and apply those lessons to ALL similar situations

**If you forget these principles, you will frustrate this user immensely. Read this section at the start of every session.**

---

## 🚨 CORE OPERATIONAL INFORMATION

### Important Facts & Credentials
- **You never need to start or restart the servers. Ever.**

### Critical UI/UX Rules
- **NEVER USE JAVASCRIPT ALERTS** - Always use toast messages or inline error displays instead
- **NO MANUAL PAGINATION** - Never implement manual pagination (Previous/Next buttons, page numbers). Always use infinite scrolling for better UX
- Use toast notifications for all success/error feedback
- Prefer inline validation and error messages in forms
- The mock should work well on desktop. Mobile support is not required.

### 🚨 Core Development Standards (NEVER VIOLATE)

#### 1. Feature Request Management
- **RECORD EVERY REQUEST**: All feature requests, no matter how small, must be documented immediately
- **ORGANIZED TRACKING**: Keep feature lists organized, prioritized, and tidy
- **COMPLETION TRACKING**: Mark features as completed when delivered

#### 2. Test-Driven Development (MANDATORY)
- **TESTS BEFORE CODE**: Always write tests before implementing new features
- **ALL TESTS MUST PASS**: Before writing new code, ensure all existing tests pass
- **NEW TESTS MUST PASS**: New code must make the new tests pass
- **NO UNTESTED CODE**: Every feature must have corresponding test coverage

#### 3. CI/CD Pipeline (ZERO FAILURES)
- **CLEAN PIPELINE**: Maintain clean CI/CD between local → GitHub → GitHub Packages
- **NO FAILED GATES**: GitHub CI/CD integration gates must never fail
- **AUTOMATED DEPLOYMENT**: Ensure seamless deployment process
- **ROLLBACK CAPABILITY**: Always maintain ability to rollback changes

#### 4. Code Quality Standards
- **NO CLEVER CODE**: Avoid clever/complex solutions in favor of clear, understandable code
- **CONSISTENT NAMING**: Use clear, consistent naming conventions throughout
- **READABLE CODE**: Code should be self-documenting and easy to understand
- **UNIFORM PATTERNS**: Follow established patterns consistently across codebase

---

## 🔧 TECHNICAL REFERENCE

### Current Architecture
- **Frontend**: HTMX with Go HTML templates
- **Backend**: Go with standard library only (net/http)
- **Data store**: In-memory store
- **Authentication**: Not needed

---

*Last updated: January 2026*
