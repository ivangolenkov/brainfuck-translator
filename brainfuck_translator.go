package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const jsFileTemplate = `const { stdout, stdin } = require('process');

const dataBufferSize = 30000
let data = new Uint8Array(dataBufferSize);
let dataPosition = 0;

const readInput = (() => {
    async function* readByteGenerator() {
        for await (const chunk of stdin) {
            for (const b of Uint8Array.from(chunk)) {
                yield b;
            }
        }
    }
    const gen = readByteGenerator();
    return async () => { data[dataPosition] = (await gen.next()).value };
})();
const changePosition = (i) => {
    dataPosition += i
    if (dataPosition > dataBufferSize - 1) {
        dataPosition -= dataBufferSize
    }
    if (dataPosition < 0) {
        dataPosition += dataBufferSize
    }
}
const writeToOutput = () => stdout.write(String.fromCharCode(data[dataPosition]));
const appendData = (i) => data[dataPosition] += i;
const isCurrentZero = () => data[dataPosition] === 0;

(async () => {
%s
})().catch(console.error);
`

func bracketCheck(prog []byte) error {
	openBrackets := 0
	for _, v := range prog {
		switch v {
		case '[':
			openBrackets++
		case ']':
			if openBrackets < 1 {
				return errors.New(
					"unexpected closing bracket",
				)
			}
			openBrackets--
		}
	}

	if openBrackets > 0 {
		return errors.New(
			"excessive opening brackets",
		)
	}

	return nil
}

func translate(prog []byte) string {
	var (
		dataChange     int64 = 0
		positionChange int64 = 0
		nesting        uint  = 1
		resultSb       strings.Builder
		operators      = map[byte]any{'+': nil, '-': nil, '>': nil, '<': nil, '.': nil, ',': nil, '[': nil, ']': nil}
	)

	for _, v := range prog {
		if _, ok := operators[v]; !ok {
			continue
		}
		if !(v == '+' || v == '-') && dataChange != 0 {
			addNestingSpaces(nesting, &resultSb)
			resultSb.WriteString(fmt.Sprintf("appendData(%d);\n", dataChange))
			dataChange = 0
		}
		if !(v == '>' || v == '<') && positionChange != 0 {
			addNestingSpaces(nesting, &resultSb)
			resultSb.WriteString(fmt.Sprintf("changePosition(%d);\n", positionChange))
			positionChange = 0
		}
		switch v {
		case '+':
			dataChange++
		case '-':
			dataChange--
		case '>':
			positionChange++
		case '<':
			positionChange--
		case '.':
			addNestingSpaces(nesting, &resultSb)
			resultSb.WriteString("await writeToOutput();\n")
		case ',':
			addNestingSpaces(nesting, &resultSb)
			resultSb.WriteString("await readInput();\n")
		case '[':
			addNestingSpaces(nesting, &resultSb)
			resultSb.WriteString("while (!isCurrentZero()) {\n")
			nesting++
		case ']':
			nesting--
			addNestingSpaces(nesting, &resultSb)
			resultSb.WriteString("}\n")
		}
	}

	return fmt.Sprintf(jsFileTemplate, resultSb.String())
}

func addNestingSpaces(nestingCount uint, sb *strings.Builder) {
	var i uint = 0
	for ; i < nestingCount; i++ {
		sb.WriteString("    ")
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [file.bf]\n", os.Args[0])
		os.Exit(1)
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	prog, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	err = bracketCheck(prog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	res := translate(prog)

	if len(os.Args) >= 3 {
		err = ioutil.WriteFile(os.Args[2], []byte(res), 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println(res)
	}
}
