# Brainfuck-translator

## Basic usage

Compile brainfuck code and write result to file:

    go run brainfuck_translator.go <path to .b file> <path to output.js file>

Compile brainfuck code and write result to stdout:

    go run brainfuck_translator.go <path to .b file>

Run brainfuck code with Node.js engine:

    node -e "$(go run brainfuck_translator.go <path to .b file>)"