# toy_grep (Toy Grep Implementation in Go)

This project is a **toy implementation of `grep -E`** written in Go.
It provides a minimal regular expression parser and matcher that can recognize and process a subset of regex features such as `.` (wildcard), `+`, `?`, `^`, `$`, character classes (`[abc]`, `[^abc]`), groups, and alternations.

---

## Features

* **Regex Anchors**: Supports `^` (start of line) and `$` (end of line).
* **Wildcards**: `.` matches any single character, `.+` matches one or more.
* **Quantifiers**:

  * `+` → one or more
  * `?` → zero or one
* **Character Classes**: e.g., `[abc]`, `[^0-9]`.
* **Escapes**:

  * `\d` → digit
  * `\w` → alphanumeric/underscore
  * `\\d` or `\\w` → literal `\d` or `\w`.
* **Grouping and Alternation**:

  * `(abc)` → group
  * `(a|b|c)` → alternation
  * `(ab)+`, `(a|b)?` → quantified groups and alternations.


---

## Usage

### Build and run

You can use it like `grep -E`:

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

* `0` → Pattern matched successfully
* `1` → No match found
* `2` → Error in execution (invalid usage, parse/match error, etc.)

---

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
```

---

## Implementation Notes

* The **parser** (`parser.go`) breaks the input regex into a linked list of tokens.
* The **matcher** (`matcher.go`) processes these tokens against input text with backtracking for `+` and `.+`.
* Supports **custom internal encodings** like:

  * `ALT:...`, `ALT+...`, `ALT?...`
  * `GRP:...`, `GRP+...`, `GRP?...`
    These are used internally to represent parsed groups/alternations.

---

## Limitations

* Not a full regex engine. Only supports a subset of features.
* Performance is not optimized — uses recursive/backtracking approach.
* Primarily educational, not production-ready.

---

## 📖 Inspiration

This project is inspired by `grep -E` and is designed as a **learning exercise** for:

* Implementing parsing and backtracking in Go
* Understanding regex internals
* Exploring linked-list based pattern representation


