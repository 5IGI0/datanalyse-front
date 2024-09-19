package main

import (
	"bytes"
	"strings"

	gotextsanitizer "github.com/5IGI0/go-text-sanitizer"
	"github.com/Masterminds/squirrel"
	"golang.org/x/net/idna"
)

func AssertError(err error) {
	if err != nil {
		panic(err)
	}
}

func GetColumn(table string, colname string) ColumnInfo {
	t := GlobalContext.Tables[table]

	if t.tableName == "" {
		return ColumnInfo{}
	}

	return t.columns[colname]
}

func SQLEscapeStringLike(str string) string {
	result := bytes.NewBufferString("")

	for _, r := range str {
		if r == '%' || r == '_' || r == '\\' {
			result.WriteByte('\\')
		}
		result.WriteRune(r)
	}

	return result.String()
}

func SanitizeText(input string) string {
	san, err := gotextsanitizer.Unidecode(input)
	AssertError(err)
	return san
}

// TODO: escape chars
func GenBidirectLike(prefix []rune, suffix []rune) string {
	var buffer bytes.Buffer

	for i := 0; i < len(prefix) || i < len(suffix); i++ {
		if i >= len(prefix) {
			buffer.WriteByte('_')
		} else {
			buffer.WriteRune(prefix[i])
		}

		if i >= len(suffix) {
			buffer.WriteByte('_')
		} else {
			buffer.WriteRune(suffix[len(suffix)-i-1])
		}
	}

	buffer.WriteByte('%')
	return buffer.String()
}

func ReverseLike(table string, column string, like string) squirrel.Like {
	rcol := GlobalContext.Tables[table].reverse_columns[column]
	if rcol == "" {
		return squirrel.Like{column: like}
	}

	reverse_buff := reverse_str([]rune(like))
	// _\ -> \_
	for i := len(reverse_buff) - 1; i > 0; i-- {
		if reverse_buff[i] == '\\' {
			tmp := reverse_buff[i-1]
			reverse_buff[i-1] = '\\'
			reverse_buff[i] = tmp
			i--
		}
	}

	return squirrel.Like{rcol: string(reverse_buff)}
}

func reverse_str(s []rune) []rune {
	reversed := make([]rune, len(s))

	for i := 0; i < len(s); i++ {
		reversed[len(s)-i-1] = s[i]
	}

	return reversed
}

func IsNumeric(input string) bool {
	for _, r := range input {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(input) != 0
}

func ReverseDomain(domain string) string {
	domain, _ = idna.ToASCII(domain)

	return string(reverse_str([]rune(strings.ToLower(domain))))
}
