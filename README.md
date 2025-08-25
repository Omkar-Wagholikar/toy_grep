# toy_grep (Toy Grep Implementation in Go)

This project is a **toy implementation of grep -E** written in Go. It provides a regular expression parser and matcher that can recognize and process a subset of regex features such as . (wildcard), +, ?, ^, $, character classes ([abc], [^abc]), groups, alternations, file searches (single and multiple) as well as recursive directory searches.

## Features

* **Regex Anchors**: Supports `^` (start of line) and `$` (end of line).
* **Wildcards**: `.` matches any single character, + matches one or more.
* **Quantifiers**:
   * `\+` → one or more
   * `?` → zero or one
* **Character Classes**: e.g., `[abc]`, `[^0-9]`, `[b-w]`
* **Escapes**:
   * `\d` → digit
   * `\w` → alphanumeric/underscore
   * `\\d` or `\\w` → literal `\d` or `\w`.
* **Grouping and Alternation**:
  * `(abc)` → group
  * `(a|b|c)` → alternation
  * `(ab)+`, `(a|b)?` → quantified groups and alternations.
  * `(a|b|c)*`  → Combined Groups

## Usage

### Build and run
Usage is similar to grep -E:
Build and run the project using the script written in toy_grep.sh 
```bash
echo -n "I see 1 cat, 2 dogs and 3 cows" | ./toy_grep.sh -E "^I see"
```
Output:
```
Successful execution
```
If the pattern does not match:
```
Error, exit with 1
```

### Exit Codes
* 0 → Pattern matched successfully
* 1 → No match found
* 2 → Error in execution (invalid parameters, improper usage, parse/match error, etc.)

## Examples

```bash
# Match start of line
echo "testcase" | ./toy_grep.sh -E "^test"

# Match digit
echo "abc123" | ./toy_grep.sh -E "\d+"

# Match character class
echo "cat" | ./toy_grep.sh -E "c[a-z]t"

# Match alternation
echo "dog" | ./toy_grep.sh -E "(cat|dog)"

# Match optional
echo "color" | ./toy_grep.sh -E "colou?r"

# Match file
./toy_grep.sh -E colo?r file.txt

# Match Multiple Files
./toy_grep.sh -E colo?r file1.txt file2.txt

# Recursively search multiple files in a directory (and sub directories)
./toy_grep.sh -r -E color dir/
```

## Implementation Details

### Folder Structure

```
.
├── app/
│   └── main.go
└── internal/
    ├── directoryWalk/
    │   └── directoryWalker.go
    ├── fileSearch/
    │   └── fileMatcher.go
    ├── parser/
    │   └── parser.go
    └── matcher/
        ├── match.go
        ├── alternationMatcher.go
        ├── baseMatchingFunctions.go
        ├── groupMatchers.go
        └── predicateFunctions.go
```

### Architecture Overview

The implementation consists of three main components:

1. **Parser** (`internal/parser/parser.go`) - Tokenizes and tags regex patterns
2. **Matcher** (`internal/matcher/matcher.go`) - Executes pattern matching with backtracking
3. **File Matcher** (`internal/fileSearch/filematcher.go`) - Searches a given array of files, line by line for a pattern match  
4. **Directory Walker** (`internal/directoryWalk/directorywalker.go`) - Used to walk a search a directory (including sub directories) to match a given pattern
3. **Pattern Cache** - Optimizes repeated parsing operations

### Pattern Parsing and Tagging System

The parser uses an intelligent tagging system to convert complex regex constructs into manageable internal representations:

#### Internal Pattern Tags

The parser transforms user-input patterns into tagged internal formats for efficient processing:

**Group Tags:**
```go
Input: (abc)
Tagged: GRP:abc

Input: (abc)+  
Tagged: GRP+:abc

Input: (abc)?
Tagged: GRP?:abc
```

**Alternation Tags:**
```go
Input: (cat|dog|cow)
Tagged: ALT:cat|dog|cow

Input: (cat|dog)+
Tagged: ALT+:cat|dog  

Input: (red|blue)?
Tagged: ALT?:red|blue
```

**Quantifier Tags:**
```go
Input: a+
Tagged: +a

Input: \d?  
Tagged: ?\d

Input: .+
Tagged: .+ (special case)
```

#### Parser Workflow

1. **Tokenization**: Break input regex into atomic units
2. **Grouping Detection**: Identify parenthesized expressions
3. **Quantifier Processing**: Attach quantifiers to preceding elements
4. **Tagging**: Apply internal tags for efficient matching
5. **Linked List Construction**: Build processing chain

**Example Parse Flow:**
```go
Input:  "^I see (\d (cat|dog|cow)(, | and )?)+$"
```

Step 1: Tokenize

```go
["^", "I see ", "(", "\d", " ", "(", "cat", "|", "dog", "|", "cow", ")", "(", ", ", "|", " and ", ")", "?", ")", "+", "$"]
```

Step 2: Process Groups and Quantifiers  
```go
["^I see ", "GRP+:\d ALT:cat|dog|cow GRP?:, | and ", "$"]
```

Step 3: Build Linked List of tags
```go
- ^I see
- GRP+:\d 
- ALT:cat|dog|cow 
- GRP?:, | and  
- $
```

#### Match Position Tracking

The engine maintains a match state that tracks:
- Current position in input text
- Current position in pattern list
- Backtrack stack for quantified expressions
- Match boundaries for anchors (^ and $)

### Performance Characteristics

**Time Complexity:**
- Best case: O(n) for simple literal matches
- Average case: O(n*m) where n=input length, m=pattern complexity  
- Worst case: O(2^n) for pathological backtracking cases

**Space Complexity:**
- O(m) for pattern storage and parsing
- O(k) for backtrack stack where k=quantifier depth
- O(c) for pattern cache where c=cache size limit

**Memory Optimizations:**
- Reuse parsed pattern structures
- Limit backtrack depth to prevent stack overflow
- Use object pooling for frequent allocations

## Limitations

* Not a full regex engine. Only supports a subset of features.
* Performance is not optimized for production use — uses a recursive/backtracking approach.
* No support for lookahead/lookbehind assertions.
* Unicode support is basic (no character class shortcuts like \p{L}).
* Primarily educational, not production-ready.

## Inspiration

This project is inspired by grep -E and is designed as a **learning exercise** for:
* Implementing parsing and backtracking algorithms in Go
* Understanding regex engine internals and optimization techniques
* Exploring pattern caching and memoization strategies
* Building maintainable, testable regex processing systems