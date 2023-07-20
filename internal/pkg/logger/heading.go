package logger

import "fmt"

func PrintHeading(content string) {
	println()

	fmt.Printf(
		"%s%s                        %s\n",
		Black,
		WhiteBackground,
		Reset,
	)

	fmt.Printf(
		"%s%s  %s  %s\n",
		Black,
		WhiteBackground,
		content,
		Reset,
	)

	fmt.Printf(
		"%s%s                        %s\n",
		Black,
		WhiteBackground,
		Reset,
	)

	println()
}
