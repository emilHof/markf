<h1 align="center">markf</h1>
<h4 align="center">Markdown to PDF renderer with macro support</h4>

---

> **NOTE** This project is still in development and is not ready for production use.

## Installation

Latest release:
```
go install github.com/OutboundSpade/markf@latest
```
Main branch:
```
go install github.com/OutboundSpade/markf@main
```

## Usage

### CLI
```
Usage markf [options] <input file>
  -allow-unsafe
    - allow unsafe macros
  -d- enable debug logging
  -o string
    - output file
  -p- print output to stdout
```
if `-o` is not specified, the output will be saved to the same directory as the input file with the same name but with the extension `.pdf`.

supported output formats (based on the extension):
- pdf
- markdown
### Custom HTML Elements

Supported HTML elements

- `<color [r],[g],[b]>` or `<color [color option]>`  - set the color of the text (see [color options](#color-options))
- `<pagebreak>` - insert a page break
- `<center>` - center the element (this only works with text & images)

#### Color Options

- `red`
- `orange`
- `yellow`
- `green`
- `blue`
- `purple`
- `white`
- `black`

## Macros

### Syntax

Macros are defined using the following syntax:

```
#!(macro-name [arg1] [arg2] [arg3] ...)
```

Lists are defined using the following syntax:

```
item1|item2|item3...
```

Characters `'`, `"`, \`, and sets of `()` are all escaped and are inteded to combine parameters with spaces into a single parameter.
eg. `#!(macro-name "this is a parameter")` will be parsed having the parameters:
- #0 - `macro-name`
- #1 - `this is a parameter`
#### Delayed Evaluation
Sets of `{}` create a "delayed evaluation" block. This means that the contents of the block will not be evaluated until the next evaluation cycle. This is useful for macros that take a body as a parameter.

eg. `#!(macro-name {#!(macro-name2)})` will be parsed having the parameters:
- #0 - `macro-name`
- #1 - `#!(macro-name2)`

whereas `#!(macro-name #!(macro-name2))` will be parsed having the parameters:

- #0 - `macro-name`
- #...- the result of `#!(macro-name2)`

#### Escape Characters
- `\n` - newline
- `\t` - tab

Please note that escape characters are not escaped until after evaluating all macros.
### Built-in Macros

- `var`: Set or get a variable 

```Usage: var <varname> ```

```Usage: var <varname> = <value...>```


- `list`: Converts anything given to it into a list  (delimited by spaces or newlines)

```Usage: list <items...>```


- `trim`: Trims a list

``` Usage: trim <from> <to> <list>```

`from` - (inclusive)

`to` - (exclusive) if `<to>` is < 0, it will include the rest of the list

- `foreach`: Loops through a list and executes a macro for each item

```Usage: foreach <varname> in <list> <body>```

You'll likely want to surround the body in curly braces to prevent the macro from being executed prematurely (see [Delayed Evaluation](#Delayed-Evaluation))

#### Unsafe Macros

> You must use the `-allow-unsafe` flag to use these macros

- `exec`: Executes a command and returns the output

```Usage: exec <command...>```

- `exec-screenshot`: Executes a command and returns a screenshot of the output

```Usage: exec-screenshot <command...>```

- `file-read`: Reads a file and returns the contents

```Usage: file-read <file>```

### Custom Macros

Macros will be searched for in the following locations:
- `./.markf-macros/`
- `~/.markf-macros/`
- a directory specified by the `MARKF_MACROS` environment variable

The macro name is the name of the file without the extension. eg. `test.md` will be called with `#!(test)`.

Custom macros can make use of the built-in & external macros.

Parameters that are passed to the macro can be read using `#$<num>` for a specific parameter or `#$...` for all parameters in list form. Parameters are 0-indexed with `#$0` being the macro name.

## Support

### Markdown Support

markf supports the following markdown elements:

- Heading
- Paragraph
- Lists
- Code Blocks
- Inline Code Blocks
- Links
- Italic
- Bold
- Horizontal Line
- Images (only png)
  - From local files, URLs, and base64 encoded data
- Text
- HTML Elements (only custom ones)
