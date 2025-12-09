package utils

import (
	"github.com/manifoldco/promptui"
)

func SelectFromOptions(options []string, label string) (string, error) {
	prompt := promptui.Select{
		Label: label,
		Items: options,
		Size:  5,
	}

	_, result, err := prompt.Run()
	return result, err
}
