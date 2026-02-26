#!/usr/bin/env bash

# HUMAN USER ONLY SCRIPT - LLM AND AGENT SHALL NEVER RUN IT


# checkpoint.sh
# Auto-checkpoint with an AI-generated, context-aware commit message.
# Extended to auto-update CLAUDE.md and CLAUDELET.md documentation.

# REQUIREMENTS:
#   - Claude CLI installed (claude -p "prompt")
#   - GIT repo already initialized
#   - jq installed (for parsing Claude streaming output)

# =============================================================================
# CONFIGURATION
# =============================================================================
# (No configuration variables currently needed)

# =============================================================================
# STEP 0: Arguments and Setup
# =============================================================================

# Check for --test or -t argument (test verbose output)
if [ "$1" = "--test" ] || [ "$1" = "-t" ]; then
  echo "=== Testing Claude verbose output ==="
  echo

  STREAM_FILE="$(mktemp)"

  # Multi-step task that requires tool use to trigger verbose reasoning output
  PROMPT="Read the file ./main.go and tell me what package it belongs to. Be brief."

  echo "[SCRIPT] Running claude..."
  echo

  # Run Claude and capture JSON stream (--verbose is required with stream-json)
  claude -p "$PROMPT" --output-format stream-json --verbose > "$STREAM_FILE"

  echo "=== Agent Reasoning Steps ==="
  # Extract all reasoning text and tool uses from the stream
  while IFS= read -r line; do
    # Extract thinking text
    text=$(echo "$line" | jq -r 'select(.type == "assistant") | .message.content[]? | select(.type == "text") | .text // empty' 2>/dev/null)
    # Only show short text (reasoning), skip long output (file content)
    [ -n "$text" ] && [ ${#text} -lt 200 ] && echo "  [thinking] $text"

    # Extract tool use
    tool=$(echo "$line" | jq -r 'select(.type == "assistant") | .message.content[]? | select(.type == "tool_use") | .name // empty' 2>/dev/null)
    [ -n "$tool" ] && echo "  [action] Using tool: $tool"
  done < "$STREAM_FILE"

  echo
  echo "=== Final Result ==="
  jq -rs '[.[] | select(.type == "result") | .result] | .[0] // "No result found"' "$STREAM_FILE"

  rm -f "$STREAM_FILE"
  exit 0
fi

# Check for --list or -l argument
if [ "$1" = "--list" ] || [ "$1" = "-l" ]; then
  echo "Recent commits:"
  echo
  git log --oneline --decorate --graph --all | tail -n 10
  exit 0
fi

# Check for --dry-run or -d argument (preview what would be updated)
if [ "$1" = "--dry-run" ] || [ "$1" = "-d" ]; then
  echo "Dry run: showing what would be updated"
  echo

  echo "Changed files:"
  git status --short

  # Get all changed files
  CHANGED_FILES="$(git diff --name-only HEAD 2>/dev/null; git diff --name-only --cached 2>/dev/null)"

  # Check if main.go or packages/**/*.go changed (triggers CLAUDE.md update)
  CLAUDE_TRIGGER="$(echo "$CHANGED_FILES" | grep -E '^(main\.go|packages/.*\.go)$')"

  # Get changed packages (for CLAUDELET updates)
  CHANGED_PKGS="$(echo "$CHANGED_FILES" | grep '^packages/' | cut -d'/' -f2 | sort -u)"

  echo
  echo "Documentation updates:"

  # Check CLAUDE.md - only if main.go or packages/*.go changed
  if [ -f "CLAUDE.md" ]; then
    if [ -n "$CLAUDE_TRIGGER" ]; then
      echo "  CLAUDE.md: WILL BE UPDATED"
    else
      echo "  CLAUDE.md: No update needed"
    fi
  else
    echo "  CLAUDE.md: NOT FOUND (skipped)"
  fi

  if [ -z "$CHANGED_PKGS" ]; then
    echo "  CLAUDELET.md: No package changes"
  else
    for pkg_name in $CHANGED_PKGS; do
      claudelet="packages/$pkg_name/CLAUDELET.md"
      if [ -f "$claudelet" ]; then
        echo "  $claudelet: WILL BE UPDATED"
      else
        echo "  $pkg_name: NO CLAUDELET.md (skipped)"
      fi
    done
  fi

  exit 0
fi

# Start timer
START_TIME=$(date +%s)

# Parse arguments
SKIP_DOCS=false
FORCE_DOCS=false
USER_COMMIT_MESSAGE=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-docs)
      SKIP_DOCS=true
      shift
      ;;
    --force|-f)
      FORCE_DOCS=true
      shift
      ;;
    *)
      USER_COMMIT_MESSAGE="$1"
      shift
      ;;
  esac
done

DATE_SHORT="$(date +%Y%m%d-%H%M%S)"
BRANCH="$(git rev-parse --abbrev-ref HEAD)"

# =============================================================================
# SHOW TODO LIST (what will be updated)
# =============================================================================
if [ "$SKIP_DOCS" = false ]; then
  echo ">>> Documentation files to update:"

  # Get all changed files
  CHANGED_FILES="$(git diff --name-only HEAD 2>/dev/null; git diff --name-only --cached 2>/dev/null)"

  # Check CLAUDE.md
  if [ "$FORCE_DOCS" = true ]; then
    echo "    CLAUDE.md (force mode)"
  else
    CLAUDE_TRIGGER="$(echo "$CHANGED_FILES" | grep -E '^(main\.go|packages/.*\.go)$')"
    if [ -n "$CLAUDE_TRIGGER" ]; then
      echo "    CLAUDE.md"
    fi
  fi

  # Check CLAUDELET.md files
  if [ "$FORCE_DOCS" = true ]; then
    for pkg_dir in packages/*/; do
      pkg_name="$(basename "$pkg_dir")"
      if [ -f "${pkg_dir}CLAUDELET.md" ]; then
        echo "    packages/$pkg_name/CLAUDELET.md (force mode)"
      fi
    done
  else
    CHANGED_PKGS="$(echo "$CHANGED_FILES" | grep '^packages/' | cut -d'/' -f2 | sort -u)"
    for pkg_name in $CHANGED_PKGS; do
      claudelet="packages/$pkg_name/CLAUDELET.md"
      if [ -f "$claudelet" ]; then
        echo "    $claudelet"
      fi
    done
  fi

  echo ""
fi

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

update_claude_md() {
  local CLAUDE_FILE="CLAUDE.md"

  if [ ! -f "$CLAUDE_FILE" ]; then
    echo "    No CLAUDE.md found, skipping"
    return
  fi

  # Only update if main.go or packages/**/*.go changed (unless --force)
  if [ "$FORCE_DOCS" != true ]; then
    local CHANGED_FILES="$(git diff --name-only HEAD 2>/dev/null; git diff --name-only --cached 2>/dev/null)"
    local CLAUDE_TRIGGER="$(echo "$CHANGED_FILES" | grep -E '^(main\.go|packages/.*\.go)$')"
    if [ -z "$CLAUDE_TRIGGER" ]; then
      echo "    Skipped CLAUDE.md (no Go source changes)"
      return
    fi
  fi

  local CURRENT="$(cat "$CLAUDE_FILE")"
  local PKG_STRUCTURE="$(find packages -type d -maxdepth 1 2>/dev/null | sort)"
  local API_ROUTES="$(grep -rh 'http.HandleFunc' packages/api/ 2>/dev/null | head -30)"

  local PROMPT="You are updating the root CLAUDE.md documentation.

CURRENT CLAUDE.md:
$CURRENT

PACKAGE STRUCTURE:
$PKG_STRUCTURE

API ROUTES (sample):
$API_ROUTES

RECENT CODE CHANGE: $COMMIT_MSG

INSTRUCTIONS:
1. Review current structure - keep same format/style
2. Update package list if new packages added
3. Update API endpoints if routes changed
4. Keep it concise - this is a reference doc
5. Output ONLY the updated markdown content"

  local RESULT
  local STREAM_FILE
  STREAM_FILE="$(mktemp)"

  # Run Claude and capture JSON stream (--verbose is required with stream-json)
  claude -p "$PROMPT" --output-format stream-json --verbose > "$STREAM_FILE"

  # Display reasoning steps from the captured stream
  while IFS= read -r line; do
    # Extract thinking text
    text=$(echo "$line" | jq -r 'select(.type == "assistant") | .message.content[]? | select(.type == "text") | .text // empty' 2>/dev/null)
    # Only show short text (reasoning), skip long output (file content)
    [ -n "$text" ] && [ ${#text} -lt 200 ] && echo "      [claude] $text"

    # Extract tool use
    tool=$(echo "$line" | jq -r 'select(.type == "assistant") | .message.content[]? | select(.type == "tool_use") | .name // empty' 2>/dev/null)
    [ -n "$tool" ] && echo "      [claude] Using tool: $tool"
  done < "$STREAM_FILE"

  # Extract final result from captured stream
  RESULT="$(jq -rs '[.[] | select(.type == "result") | .result] | .[0] // ""' "$STREAM_FILE")"
  rm -f "$STREAM_FILE"

  if [ -n "$RESULT" ]; then
    echo "$RESULT" > "$CLAUDE_FILE"
    echo "    Updated CLAUDE.md (${#RESULT} chars)"
  else
    echo "    Skipped CLAUDE.md (empty response)"
  fi
}

update_claudelet_files() {
  local CHANGED_PKGS

  if [ "$FORCE_DOCS" = true ]; then
    # Force mode: update ALL packages with CLAUDELET.md
    echo "    Force mode: updating all packages"
    CHANGED_PKGS="$(ls -d packages/*/ 2>/dev/null | xargs -n1 basename)"
  else
    # Normal mode: only changed packages from git diff
    CHANGED_PKGS="$(git diff --name-only HEAD 2>/dev/null; git diff --name-only --cached 2>/dev/null)"
    CHANGED_PKGS="$(echo "$CHANGED_PKGS" | grep '^packages/' | cut -d'/' -f2 | sort -u)"

    if [ -z "$CHANGED_PKGS" ]; then
      echo "    No packages changed, skipping all CLAUDELET updates"
      return
    fi

    echo "    Changed packages: $CHANGED_PKGS"
  fi

  for pkg_name in $CHANGED_PKGS; do
    local pkg_dir="packages/$pkg_name"
    local claudelet="${pkg_dir}/CLAUDELET.md"

    if [ ! -d "$pkg_dir" ]; then
      echo "    Skipping $pkg_name (directory not found)"
      continue
    fi

    if [ ! -f "$claudelet" ]; then
      echo "    Skipping $pkg_name (no CLAUDELET.md)"
      continue
    fi

    update_single_claudelet "$pkg_dir" "$pkg_name" "$claudelet"
  done
}

update_single_claudelet() {
  local pkg_dir="$1"
  local pkg_name="$2"
  local claudelet="$3"

  echo "    Processing $pkg_name..."

  local CURRENT="$(cat "$claudelet")"

  # Gather Go source (first 150 lines of each .go file)
  local GO_SRC=""
  while IFS= read -r -d '' gofile; do
    GO_SRC+="
=== $(basename "$gofile") ===
$(head -150 "$gofile")
"
  done < <(find "$pkg_dir" -name "*.go" -print0 2>/dev/null)

  local PROMPT="You are updating CLAUDELET.md for package: $pkg_name

CURRENT CLAUDELET.md:
$CURRENT

PACKAGE SOURCE FILES:
$GO_SRC

INSTRUCTIONS:
1. Keep same format and section structure
2. Update 'Key Files' with current files and line counts
3. Update 'Exports' if public functions changed
4. Update 'Purpose' only if functionality changed significantly
5. Output ONLY the updated markdown content"

  local RESULT
  local STREAM_FILE
  STREAM_FILE="$(mktemp)"

  # Run Claude and capture JSON stream (--verbose is required with stream-json)
  claude -p "$PROMPT" --output-format stream-json --verbose > "$STREAM_FILE"

  # Display reasoning steps from the captured stream
  while IFS= read -r line; do
    # Extract thinking text
    text=$(echo "$line" | jq -r 'select(.type == "assistant") | .message.content[]? | select(.type == "text") | .text // empty' 2>/dev/null)
    # Only show short text (reasoning), skip long output (file content)
    [ -n "$text" ] && [ ${#text} -lt 200 ] && echo "      [claude] $text"

    # Extract tool use
    tool=$(echo "$line" | jq -r 'select(.type == "assistant") | .message.content[]? | select(.type == "tool_use") | .name // empty' 2>/dev/null)
    [ -n "$tool" ] && echo "      [claude] Using tool: $tool"
  done < "$STREAM_FILE"

  # Extract final result from captured stream
  RESULT="$(jq -rs '[.[] | select(.type == "result") | .result] | .[0] // ""' "$STREAM_FILE")"
  rm -f "$STREAM_FILE"

  if [ -n "$RESULT" ]; then
    echo "$RESULT" > "$claudelet"
    echo "    Updated $claudelet"
  else
    echo "    Skipped $pkg_name (empty response)"
  fi
}

# =============================================================================
# STEP 1: GENERATE COMMIT MESSAGE (BEFORE doc updates)
# =============================================================================
echo ">>> Step 1: Generating commit message..."

DIFF="$(git diff --cached; git diff)"
STATUS="$(git status --short)"
RECENT_COMMITS="$(git log --oneline --decorate --graph --all | tail -n 10)"

# Build prompt for Claude
COMMIT_PROMPT=$(cat <<EOF
INSTRUCTIONS:
You are an expert developer. Generate a SHORT but relevant descriptive git commit message
(1 line max) based strictly on the following data.
You create a commit message that take the application goal and its commit history to craft relevant commit messages.
This relevant commit explains what exactly have been done using real resources and keywords.

APPLICATION GOAL:
Generic Go service application

COMMIT HISTORY:
$RECENT_COMMITS

Branch: $BRANCH

Git Status:
$STATUS

Git Diff:
$DIFF

Important:
- Summarize the intent of the changes.
- DO NOT mention files individually unless necessary.
- DO NOT mention CLAUDE.md or CLAUDELET.md documentation updates.
- DO NOT say "updated files" or "misc changes".
- Produce ONLY the commit message text. No commentary.
EOF
)

COMMIT_MSG="$(claude -p "$COMMIT_PROMPT")"

# Fallback if Claude returns empty
if [ -z "$COMMIT_MSG" ]; then
  COMMIT_MSG="auto-checkpoint-$BRANCH-$DATE_SHORT"
fi

# Prepend user message if provided
if [ -n "$USER_COMMIT_MESSAGE" ]; then
  COMMIT_MSG="$USER_COMMIT_MESSAGE: $COMMIT_MSG"
fi

echo ">>> Commit message: $COMMIT_MSG"

# =============================================================================
# STEP 2: UPDATE CLAUDE.md (if not skipped)
# =============================================================================
if [ "$SKIP_DOCS" = false ]; then
  echo ">>> Step 2: Updating CLAUDE.md..."
  update_claude_md
else
  echo ">>> Step 2: Skipped (--skip-docs)"
fi

# =============================================================================
# STEP 3: UPDATE CLAUDELET.md FILES (if not skipped)
# =============================================================================
if [ "$SKIP_DOCS" = false ]; then
  echo ">>> Step 3: Updating CLAUDELET.md files..."
  update_claudelet_files
else
  echo ">>> Step 3: Skipped (--skip-docs)"
fi

# =============================================================================
# STEP 4: COMMIT
# =============================================================================
echo ">>> Step 4: Committing..."

git add -A
git commit -m "$COMMIT_MSG--$BRANCH-$DATE_SHORT"

# =============================================================================
# STEP 5: BUMP API_VERSION (with commit hash)
# =============================================================================
echo ">>> Step 5: Bumping API_VERSION..."

if [ -f "API_VERSION" ]; then
  CURRENT_VERSION="$(cat API_VERSION)"
  # Strip any existing build metadata suffix
  VERSION_BASE="${CURRENT_VERSION%%+*}"
  # Extract major.minor and patch
  MAJOR_MINOR="${VERSION_BASE%.*}"
  PATCH="${VERSION_BASE##*.}"
  # Increment patch
  NEW_PATCH=$((PATCH + 1))
  # Get short commit hash
  COMMIT_HASH="$(git rev-parse --short HEAD)"
  NEW_VERSION="${MAJOR_MINOR}.${NEW_PATCH}+${COMMIT_HASH}"
  echo "$NEW_VERSION" > API_VERSION
  echo "    $CURRENT_VERSION -> $NEW_VERSION"
  # Amend the commit to include the version bump
  git add API_VERSION
  git commit --amend --no-edit
else
  echo "    API_VERSION file not found, skipping"
fi

echo ""
echo ""
echo "CHECKPOINT CREATED:"
echo "$COMMIT_MSG"

# End timer and calculate duration
END_TIME=$(date +%s)
ELAPSED=$((END_TIME - START_TIME))
MINUTES=$((ELAPSED / 60))
SECONDS=$((ELAPSED % 60))
echo ""
echo "Code checkpoint completed in: ${MINUTES}m${SECONDS}s"
git status
