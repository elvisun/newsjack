package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

type setupChoice struct {
	Value    string
	Label    string
	Hint     string
	Selected bool
}

func (w *setupWizard) selectMulti(question string, choices []setupChoice) ([]string, bool) {
	stdinFile, stdoutFile, ok := interactiveFiles(w.stdin, w.stdout)
	if !ok || w.assumeYes {
		var defaults []string
		for _, choice := range choices {
			if choice.Selected {
				defaults = append(defaults, choice.Value)
			}
		}
		return defaults, false
	}
	state, err := term.MakeRaw(int(stdinFile.Fd()))
	if err != nil {
		return nil, false
	}
	defer func() { _ = term.Restore(int(stdinFile.Fd()), state) }()

	cursor := firstSelectedChoice(choices)
	lines := 0
	for {
		if lines > 0 {
			clearMenu(stdoutFile, lines)
		}
		lines = renderChoiceMenu(stdoutFile, question, choices, cursor, true)
		key, err := readTerminalKey(stdinFile)
		if err != nil {
			return nil, false
		}
		switch key {
		case "up":
			cursor = (cursor - 1 + len(choices)) % len(choices)
		case "down":
			cursor = (cursor + 1) % len(choices)
		case "space":
			choices[cursor].Selected = !choices[cursor].Selected
		case "enter":
			var values []string
			var labels []string
			for _, choice := range choices {
				if choice.Selected {
					values = append(values, choice.Value)
					labels = append(labels, choice.Label)
				}
			}
			if len(values) == 0 {
				fmt.Fprint(stdoutFile, "\a")
				continue
			}
			clearMenu(stdoutFile, lines)
			fmt.Fprintf(stdoutFile, "%s %s\n", uiQuestion(stdoutFile), question)
			uiKV(stdoutFile, "selected", strings.Join(labels, ", "))
			w.interactive = true
			return values, true
		case "ctrl-c":
			clearMenu(stdoutFile, lines)
			return nil, false
		}
	}
}

func (w *setupWizard) selectSingle(question string, choices []setupChoice, defaultValue string) (string, bool) {
	stdinFile, stdoutFile, ok := interactiveFiles(w.stdin, w.stdout)
	if !ok || w.assumeYes {
		return defaultValue, false
	}
	state, err := term.MakeRaw(int(stdinFile.Fd()))
	if err != nil {
		return "", false
	}
	defer func() { _ = term.Restore(int(stdinFile.Fd()), state) }()

	cursor := choiceIndex(choices, defaultValue)
	lines := 0
	for {
		if lines > 0 {
			clearMenu(stdoutFile, lines)
		}
		lines = renderChoiceMenu(stdoutFile, question, choices, cursor, false)
		key, err := readTerminalKey(stdinFile)
		if err != nil {
			return "", false
		}
		switch key {
		case "up":
			cursor = (cursor - 1 + len(choices)) % len(choices)
		case "down":
			cursor = (cursor + 1) % len(choices)
		case "space", "enter":
			clearMenu(stdoutFile, lines)
			fmt.Fprintf(stdoutFile, "%s %s\n", uiQuestion(stdoutFile), question)
			uiKV(stdoutFile, "selected", choices[cursor].Label)
			w.interactive = true
			return choices[cursor].Value, true
		case "ctrl-c":
			clearMenu(stdoutFile, lines)
			return "", false
		}
	}
}

func interactiveFiles(stdin io.Reader, stdout io.Writer) (*os.File, *os.File, bool) {
	stdinFile, stdinOK := stdin.(*os.File)
	stdoutFile, stdoutOK := stdout.(*os.File)
	if !stdinOK || !stdoutOK {
		return nil, nil, false
	}
	if strings.ToLower(strings.TrimSpace(os.Getenv("TERM"))) == "dumb" {
		return nil, nil, false
	}
	if !term.IsTerminal(int(stdinFile.Fd())) || !term.IsTerminal(int(stdoutFile.Fd())) {
		return nil, nil, false
	}
	return stdinFile, stdoutFile, true
}

func firstSelectedChoice(choices []setupChoice) int {
	for i, choice := range choices {
		if choice.Selected {
			return i
		}
	}
	return 0
}

func choiceIndex(choices []setupChoice, value string) int {
	for i, choice := range choices {
		if choice.Value == value {
			return i
		}
	}
	return 0
}

func renderChoiceMenu(stdout io.Writer, question string, choices []setupChoice, cursor int, multi bool) int {
	fmt.Fprintf(stdout, "%s %s\n", uiQuestion(stdout), question)
	if multi {
		uiNote(stdout, "use up/down to move, space to toggle, enter to continue.")
	} else {
		uiNote(stdout, "use up/down to move, enter to continue.")
	}
	for i, choice := range choices {
		pointer := " "
		if i == cursor {
			pointer = ">"
		}
		marker := "( )"
		if multi {
			marker = "[ ]"
			if choice.Selected {
				marker = "[x]"
			}
		} else if i == cursor {
			marker = "(*)"
		}
		label := fmt.Sprintf("%-14s", choice.Label)
		if i == cursor {
			label = uiPaint(stdout, ansiBold, label)
		}
		hint := ""
		if choice.Hint != "" {
			hint = "  " + uiPaint(stdout, ansiDim, choice.Hint)
		}
		fmt.Fprintf(stdout, "  %s %s %s%s\n", pointer, marker, label, hint)
	}
	return len(choices) + 2
}

func clearMenu(stdout io.Writer, lines int) {
	fmt.Fprintf(stdout, "\x1b[%dA\x1b[J", lines)
}

func readTerminalKey(stdin *os.File) (string, error) {
	var b [3]byte
	if _, err := stdin.Read(b[:1]); err != nil {
		return "", err
	}
	switch b[0] {
	case 0x03:
		return "ctrl-c", nil
	case '\r', '\n':
		return "enter", nil
	case ' ', 'x':
		return "space", nil
	case 'k':
		return "up", nil
	case 'j':
		return "down", nil
	case 0x1b:
		if _, err := stdin.Read(b[1:2]); err != nil {
			return "", err
		}
		if b[1] != '[' {
			return "", nil
		}
		if _, err := stdin.Read(b[2:3]); err != nil {
			return "", err
		}
		switch b[2] {
		case 'A':
			return "up", nil
		case 'B':
			return "down", nil
		}
	}
	return "", nil
}
