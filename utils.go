package main

import (
	"math/rand"
)

func randString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func randCap(s string) string {
	output := ""
	for _, c := range s {
		if rand.Intn(2) == 0 {
			output += string(c)
		} else {
			if c >= 'a' && c <= 'z' {
				output += string(c - 'a' + 'A')
			} else if c >= 'A' && c <= 'Z' {
				output += string(c - 'A' + 'a')
			} else {
				output += string(c)
			}
		}
	}
	return output
}
