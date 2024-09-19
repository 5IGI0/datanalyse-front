package main

import "strings"

// expect international format without +
func FormatPhone(phone string) []string {

	if phone, found := strings.CutPrefix(phone, "1"); found {
		// North America
		return []string{"1" + phone, phone}
	}

	if phone, found := strings.CutPrefix(phone, "33"); found {
		// France
		return []string{
			"0" + phone,     // 0712345678
			"33" + phone,    // +33712345678
			"0033" + phone,  // 0033712345678
			"033" + phone,   // 0 (33) 712345678
			"330" + phone,   // 33 (0) 712345678
			"00330" + phone} // 0033 (0) 712345678
	}

	return []string{phone}
}
