package main

import "strings"

func eraseProfane(msg string) string {
	splitted := strings.Split(msg, " ")
	for i, word := range splitted {
		lowerWord := strings.ToLower(word)
		if lowerWord == "kerfuffle" || lowerWord == "sharbert" || lowerWord == "fornax" {
			splitted[i] = "****"
		}
	}
	return strings.Join(splitted, " ")
}
