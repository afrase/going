package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/chzyer/readline"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/manifoldco/promptui"
	"github.com/manifoldco/promptui/list"
)

// noBellStdout disabled readline's annoying terminal bell.
type noBellStdout struct{}

func (n *noBellStdout) Write(p []byte) (int, error) {
	if len(p) == 1 && p[0] == readline.CharBell {
		return 0, nil
	}
	return readline.Stdout.Write(p)
}

func (n *noBellStdout) Close() error {
	return readline.Stdout.Close()
}

const selectItemSize = 10

type Prompt interface {
	Select(label string, items []string) (value string)
	CustomSelect(label string, items interface{}, tmpl *promptui.SelectTemplates, searcher list.Searcher) (index int)
	YesNoPrompt(label string) bool
}

type Prompter struct{}

var ContainerTemplate = &promptui.SelectTemplates{
	Label:    fmt.Sprintf("%s {{ .Name }}: ", promptui.IconInitial),
	Active:   fmt.Sprintf("%s {{ .Name | underline }}", promptui.IconSelect),
	Inactive: "  {{ .Name }}",
	Selected: fmt.Sprintf(`{{ "%s" | green }} {{ .Name | faint }}`, promptui.IconGood),
}

func (p Prompter) Select(label string, items []string) string {
	prompt := promptui.Select{
		Label:             label,
		Items:             items,
		Size:              selectItemSize,
		Searcher:          fuzzyStringSearch(items),
		StartInSearchMode: true,
		Stdout:            &noBellStdout{},
	}

	_, result, err := prompt.Run()
	CheckErr(err)

	return result
}

func (p Prompter) CustomSelect(label string, items interface{}, tmpl *promptui.SelectTemplates, searcher list.Searcher) int {
	prompt := promptui.Select{
		Label:             label,
		Items:             items,
		Size:              selectItemSize,
		Templates:         tmpl,
		Searcher:          searcher,
		StartInSearchMode: true,
		Stdout:            &noBellStdout{},
	}
	i, _, err := prompt.Run()
	CheckErr(err)

	return i
}

func (p Prompter) YesNoPrompt(label string) bool {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
		Default:   "n",
		Stdout:    &noBellStdout{},
		Validate: func(s string) error {
			if len(s) == 1 && strings.Contains("YyNn", s) || len(s) == 0 {
				return nil
			}
			return errors.New("invalid input")
		},
	}

	_, err := prompt.Run()
	aborted := errors.Is(err, promptui.ErrAbort)
	if !aborted && err == nil {
		return true
	} else {
		return false
	}
}

func fuzzyStringSearch(itemsToSelect []string) func(input string, index int) bool {
	return func(input string, index int) bool {
		item := itemsToSelect[index]
		if fuzzy.MatchFold(input, item) {
			return true
		}
		return false
	}
}
