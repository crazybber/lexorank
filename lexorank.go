// Package lexorank is a simple implementation of LexoRank.
//
// LexoRank is a ranking system introduced by Atlassian JIRA.
// For details - https://www.youtube.com/watch?v=OjQv9xMoFbg
package lexorank

import (
	"fmt"
	"regexp"
)

const orderToByte = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func byteToOrder(b byte) byte {
	n := _byteToOrder(b)
	if n < 0 || n > 62 {
		panic("bad value")
	}
	return n
}

func _byteToOrder(b byte) byte {
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

// Parses a Jira (Cloud?) lexorank field, which seems to look like:
//    <bucket>|<base36>[:<base36>]

type Posn struct {
	Bucket byte
	Major  string
	Minor  string // note this includes the ":" prefix
}

var jiraRank = regexp.MustCompile(`^([012])\|([0-9a-z]+)(:[0-9a-z]*)?$`)

func (p Posn) String() string {
	return fmt.Sprintf("%d|%s%s", p.Bucket, p.Major, p.Minor)
}

func ParseJira(rank string) (Posn, bool) {
	m := jiraRank.FindStringSubmatch(rank)
	if m == nil {
		return Posn{}, false
	}
	return Posn{
		Bucket: m[1][0] - '0',
		Major:  m[2],
		Minor:  m[3],
	}, true
}

const MaxMultiRank = 10 + 26 + 26 - 1

// Ranks arranges for there to be N ranks between `prev` and `next`
// and returns them.  This is useful when re-ranking a group of
// objects together at onces.
func Ranks(n int, prev, next *Posn) ([]Posn, bool) {
	if n > MaxMultiRank {
		// can't accommodate that many all at once
		return nil, false
	}

	if prev == nil {
		prev = &Posn{
			Major: "000000",
			Minor: ":",
		}
		// if there *is* a next, adopt its bucket
		if next != nil {
			prev.Bucket = next.Bucket
		}
	}

	if next == nil {
		next = &Posn{
			Major: "zzzzzz",
			Minor: ":",
		}
		// if there *is* a prev, adopt its bucket
		if prev != nil {
			next.Bucket = prev.Bucket
		}
	}

	if prev.Major != next.Major {
		p, ok := majorRanks(n, *prev, *next)
		if ok {
			return p, true
		}
	}
	return minorRanks(n, *prev, *next)
}

func minorRanks(n int, prev, next Posn) ([]Posn, bool) {
	panic("TODO")
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func majorRanks(n int, prev, next Posn) ([]Posn, bool) {
	rank := ""
	i := 0

	majorLen := max(len(prev.Major), len(next.Major))

	for {
		prevChar := getChar(prev.Major, i, minChar)
		nextChar := getChar(next.Major, i, maxChar)

		if prevChar == nextChar {
			// common prefix
			fmt.Printf("common prefix at [%c]\n", prevChar)
			rank += string(prevChar)
			i++
			continue
		}

		midChars, ok := mids(n, prevChar, nextChar)
		if !ok {
			// we need to adjust the bounds in which we're searching for ranks
			// at this point we have an uncommon prefix, e.g.,
			//   |||
			//   005z
			//   006b
			// for the next iteration we want to use one with the most space
			// avaialble, which means going forward with
			//   0060
			//   006b
			fmt.Printf("fork in the road at [%c <> %c]\n", prevChar, nextChar)
			prevAfter := byteToOrder(getChar(prev.Major, i+1, minChar))
			nextAfter := byteToOrder(getChar(next.Major, i+1, maxChar))
			spaceAfterPrev := MaxMultiRank - prevAfter
			spaceBeforeNext := nextAfter
			fmt.Printf("   after this, PREV has order %d (space %d)\n", prevAfter, spaceAfterPrev)
			fmt.Printf("               NEXT has order %d (space %d)\n", nextAfter, spaceBeforeNext)

			if spaceAfterPrev > spaceBeforeNext {
				next.Major = next.Major[:i] + string(prevChar)
				fmt.Printf("  go forward with NEXT [%s]\n", next.Major)
				rank += string(prevChar)
			} else {
				prev.Major = prev.Major[:i] + string(nextChar)
				fmt.Printf("  go forward with PREV [%s]\n", prev.Major)
				rank += string(nextChar)
			}
			i++
			continue
		}

		if len(rank) == majorLen {
			return nil, false
		}

		out := make([]Posn, n)
		// arrange for the major parts to all be the same size
		// by attaching a trailer to newly generated major ranks
		trailer := trailer[:majorLen-1-len(rank)]

		for j, mid := range midChars {
			out[j] = Posn{
				Bucket: prev.Bucket,
				Major:  rank + string(mid) + trailer,
				Minor:  ":",
			}
		}
		return out, true
	}
}

const trailer = "UUUUUUUU"

func mids(n int, prev, next byte) ([]byte, bool) {
	prevo := byteToOrder(prev)
	nexto := byteToOrder(next)
	per := int(nexto-prevo) / (n + 1)
	if per < 1 {
		return nil, false
	}
	fmt.Printf("(%c ... %c)  is (%d ... %d)  per is %d\n",
		prev, next,
		prevo, nexto,
		per)

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
