package main

import (
	"io"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
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
		return selectedChoiceValues(choices), false
	}

	var selected []string
	prompt := &survey.MultiSelect{
		Message:     question,
		Options:     choiceLabels(choices),
		Default:     selectedChoiceLabels(choices),
		Description: choiceDescription(choices),
		PageSize:    len(choices),
		VimMode:     true,
	}
	if err := survey.AskOne(prompt, &selected, survey.WithStdio(stdinFile, stdoutFile, w.stderr), survey.WithValidator(survey.Required)); err != nil {
		return nil, false
	}
	w.interactive = true
	uiKV(stdoutFile, "selected", strings.Join(selected, ", "))
	return choiceValuesForLabels(choices, selected), true
}

func (w *setupWizard) selectSingle(question string, choices []setupChoice, defaultValue string) (string, bool) {
	stdinFile, stdoutFile, ok := interactiveFiles(w.stdin, w.stdout)
	if !ok || w.assumeYes {
		return defaultValue, false
	}

	defaultLabel := choiceLabelForValue(choices, defaultValue)
	selected := defaultLabel
	prompt := &survey.Select{
		Message:     question,
		Options:     choiceLabels(choices),
		Default:     defaultLabel,
		Description: choiceDescription(choices),
		PageSize:    len(choices),
		VimMode:     true,
	}
	if err := survey.AskOne(prompt, &selected, survey.WithStdio(stdinFile, stdoutFile, w.stderr)); err != nil {
		return "", false
	}
	w.interactive = true
	uiKV(stdoutFile, "selected", selected)
	return choiceValueForLabel(choices, selected), true
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

func choiceLabels(choices []setupChoice) []string {
	labels := make([]string, 0, len(choices))
	for _, choice := range choices {
		labels = append(labels, choice.Label)
	}
	return labels
}

func selectedChoiceLabels(choices []setupChoice) []string {
	var labels []string
	for _, choice := range choices {
		if choice.Selected {
			labels = append(labels, choice.Label)
		}
	}
	return labels
}

func selectedChoiceValues(choices []setupChoice) []string {
	var values []string
	for _, choice := range choices {
		if choice.Selected {
			values = append(values, choice.Value)
		}
	}
	return values
}

func choiceDescription(choices []setupChoice) func(string, int) string {
	return func(_ string, index int) string {
		if index < 0 || index >= len(choices) {
			return ""
		}
		return choices[index].Hint
	}
}

func choiceLabelForValue(choices []setupChoice, value string) string {
	for _, choice := range choices {
		if choice.Value == value {
			return choice.Label
		}
	}
	if len(choices) == 0 {
		return ""
	}
	return choices[0].Label
}

func choiceValueForLabel(choices []setupChoice, label string) string {
	values := choiceValuesForLabels(choices, []string{label})
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func choiceValuesForLabels(choices []setupChoice, labels []string) []string {
	selected := make(map[string]bool, len(labels))
	for _, label := range labels {
		selected[label] = true
	}
	values := make([]string, 0, len(labels))
	for _, choice := range choices {
		if selected[choice.Label] {
			values = append(values, choice.Value)
		}
	}
	return values
}
