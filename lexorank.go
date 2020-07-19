// Package lexorank is a simple implementation of LexoRank.
//
// LexoRank is a ranking system introduced by Atlassian JIRA.
// For details - https://www.youtube.com/watch?v=OjQv9xMoFbg
package lexorank

const orderToByte = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func byteToOrder(b byte) byte {
	if b >= '0' && b <= '9' {
		return b - '0'
	} else if b >= 'A' && b <= 'Z' {
		return b - 'A' + 10
	} else if b >= 'a' && b <= 'z' {
		return b - 'a' + 10 + 26
	} else if b < 'A' {
		return 0
	} else {
		return 10 + 26 + 26 - 1
	}
}

const (
	minChar = byte('0')
	maxChar = byte('z')
)

// Rank returns a new rank string between prev and next.
// ok=false if it needs to be reshuffled. e.g. same or adjacent prev, next values.
func Rank(prev, next string) (string, bool) {
	lst, ok := Ranks(1, prev, next)
	if !ok {
		return prev, false
	}
	return lst[0], true
}

const MaxMultiRank = 10 + 26 + 26 - 1

// Ranks arranges for there to be N ranks between `prev` and `next`
// and returns them.  This is useful when re-ranking a group of
// objects together at onces.
func Ranks(n int, prev, next string) ([]string, bool) {
	if n > MaxMultiRank {
		// can't accommodate that many at all
		return nil, false
	}

	if prev == "" {
		prev = string(minChar)
	}
	if next == "" {
		next = string(maxChar)
	}

	rank := ""
	i := 0

	for {
		prevChar := getChar(prev, i, minChar)
		nextChar := getChar(next, i, maxChar)

		if prevChar == nextChar {
			rank += string(prevChar)
			i++
			continue
		}

		midChars, ok := mids(n, prevChar, nextChar)
		if !ok {
			rank += string(prevChar)
			i++
			continue
		}

		out := make([]string, n)
		for j, mid := range midChars {
			out[j] = rank + string(mid)
		}
		return out, true
	}
}

func mids(n int, prev, next byte) ([]byte, bool) {
	prevo := byteToOrder(prev)
	nexto := byteToOrder(next)
	per := int(nexto-prevo) / (n + 1)
	if per < 1 {
		return nil, false
	}
	ch := make([]byte, n)
	for i := 0; i < n; i++ {
		ch[i] = orderToByte[int(prevo)+per*(i+1)]
	}
	return ch, true
}

func getChar(s string, i int, defaultChar byte) byte {
	if i >= len(s) {
		return defaultChar
	}
	return s[i]
}
