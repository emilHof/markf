<h1 style="text-align:center;">markf</h1>
<h4 style="text-align:center;">Markdown to PDF renderer with macro support</h4>

---

> **NOTE** This project is still in development and is not ready for production use.

## Installation

```
go install github.com/OutboundSpade/markf
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
### Custom HTML Elements

Supported HTML elements

- `<color [r],[g],[b]>` or `<color [color option]>`  - set the color of the text (see [color options](####color-options))
- `<pagebreak>` - insert a page break
- `<center>` - center the element

#### Color Options

- `red`
- `green`
- `blue`
- `yellow`
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

### Built-in Macros

- `var`: Set or get a variable 

```Usage: var <varname> ```
```Usage: var <varname> = <value...>```


- `list`: Converts anything given to it into a list  (delimited by spaces or newlines)

```Usage: list <items...>```


- `trim`: Trims a list

``` Usage: trim <from> <to (exclusive)> <list>```
`from` - (inclusive)
`to` - (exclusive) if `<to>` is < 0, it will include the rest of the list

- `foreach`: Loops through a list and executes a macro for each item

```Usage: foreach <varname> in <list> <body>```
> You'll likely want to surround the body in curly braces to prevent the macro from being executed prematurely

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

Parameters that are passed to the macro can be read using `#$<num>` for a specific parameter or `#$...` for all parameters in list form.

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
- Images
- Text
- HTML Elements (only custom ones)
