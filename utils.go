package main

import "strings"

func eraseProfane(msg chirp) map[string]string {
	splitted := strings.Split(msg.Body, " ")
	for i, word := range splitted {
		lowerWord := strings.ToLower(word)
		if lowerWord == "kerfuffle" || lowerWord == "sharbert" || lowerWord == "fornax" {
			splitted[i] = "****"
		}
	}
	return map[string]string{"cleaned_body": strings.Join(splitted, " ")}
}
