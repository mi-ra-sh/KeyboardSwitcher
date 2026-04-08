package main

// EN↔UA character mapping based on physical key positions.
// Same physical key produces different characters depending on the active layout.

var enToUA map[rune]rune
var uaToEN map[rune]rune

func init() {
	enToUA = make(map[rune]rune)
	uaToEN = make(map[rune]rune)

	// Lowercase
	en := "`qwertyuiop[]\\asdfghjkl;'zxcvbnm,./"
	ua := "'йцукенгшщзхїґфівапролджєячсмитьбю."

	for i, e := range en {
		u := rune([]rune(ua)[i])
		enToUA[e] = u
		uaToEN[u] = e
	}

	// Uppercase
	enUp := "~QWERTYUIOP{}|ASDFGHJKL:\"ZXCVBNM<>?"
	uaUp := "₴ЙЦУКЕНГШЩЗХЇҐФІВАПРОЛДЖЄЯЧСМИТЬБЮ,"

	for i, e := range enUp {
		u := rune([]rune(uaUp)[i])
		enToUA[e] = u
		uaToEN[u] = e
	}

	// Shift+number row differences
	enShift := "!@#$%^&*()_+"
	uaShift := "!\"№;%:?*()_+"

	for i, e := range enShift {
		u := rune([]rune(uaShift)[i])
		if e != u {
			enToUA[e] = u
			uaToEN[u] = e
		}
	}
}

// convertText detects the direction (EN→UA or UA→EN) and converts.
func convertText(text []rune) []rune {
	if len(text) == 0 {
		return text
	}

	var uaCount, enCount int
	for _, c := range text {
		if isUA(c) {
			uaCount++
		} else if isEN(c) {
			enCount++
		}
	}

	triedUA := false
	var mapping map[rune]rune
	if uaCount > enCount {
		mapping = uaToEN
	} else if enCount > uaCount {
		mapping = enToUA
	} else {
		mapping = uaToEN
		triedUA = true
	}

	result := make([]rune, len(text))
	changed := false
	for i, c := range text {
		if m, ok := mapping[c]; ok {
			result[i] = m
			changed = true
		} else {
			result[i] = c
		}
	}

	if !changed && triedUA {
		mapping = enToUA
		for i, c := range text {
			if m, ok := mapping[c]; ok {
				result[i] = m
				changed = true
			} else {
				result[i] = c
			}
		}
	}

	if !changed {
		return text
	}
	return result
}

func isUA(c rune) bool {
	return (c >= 'а' && c <= 'я') || (c >= 'А' && c <= 'Я') ||
		c == 'і' || c == 'І' || c == 'ї' || c == 'Ї' ||
		c == 'є' || c == 'Є' || c == 'ґ' || c == 'Ґ'
}

func isEN(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
