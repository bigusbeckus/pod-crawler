package logger

import (
	"fmt"
	"io"
	"strings"
)

type colorWriter struct {
	out io.Writer
}

const infoPrefix = "[ INFO  ]  "
const warnPrefix = "[ WARN  ]  "
const errorPrefix = "[ ERROR ]  "

func (cw *colorWriter) Write(p []byte) (n int, err error) {
	str := string(p)

	if strings.HasPrefix(str, infoPrefix) {
		str = strings.ReplaceAll(str, infoPrefix, fmt.Sprintf("%s%s%s", Blue, infoPrefix, Reset))
	} else if strings.HasPrefix(str, warnPrefix) {
		str = strings.ReplaceAll(str, warnPrefix, fmt.Sprintf("%s%s%s", Yellow, warnPrefix, Reset))
	} else if strings.HasPrefix(str, errorPrefix) {
		str = strings.ReplaceAll(str, errorPrefix, fmt.Sprintf("%s%s%s", Red, errorPrefix, Reset))
	}

	return cw.out.Write([]byte(str))
}
