package main

import (
	"slices"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
)

func nodeToString(str *strings.Builder, node *html.Node) {
	switch node.Type {
	case html.TextNode:
		str.WriteString(removeNonUtf8(node.Data))
	case html.ElementNode:
		switch node.Data {
		case "span":
			for nextNode := node.FirstChild; nextNode != nil; nextNode = nextNode.NextSibling {
				nodeToString(str, nextNode)
			}
		case "br":
			str.WriteByte('\n')
		case "b":
			str.WriteString("**")
			for nextNode := node.FirstChild; nextNode != nil; nextNode = nextNode.NextSibling {
				nodeToString(str, nextNode)
			}
			str.WriteString("**")
		case "i":
			str.WriteByte('_')
			for nextNode := node.FirstChild; nextNode != nil; nextNode = nextNode.NextSibling {
				nodeToString(str, nextNode)
			}
			str.WriteByte('_')
		case "a":
			targetIdx := slices.IndexFunc(node.Attr, func(a html.Attribute) bool { return a.Key == "href" })
			if targetIdx == -1 {
				str.WriteString(removeNonUtf8(node.FirstChild.Data))
				return
			}

			target := node.Attr[targetIdx].Val
			if strings.HasPrefix(target, "http") || !strings.Contains(target, ".asp") {
				str.WriteString("[")
				str.WriteString(removeNonUtf8(node.FirstChild.Data))
				str.WriteString("](")
				str.WriteString(target)
				str.WriteString(")")
			} else {
				str.WriteString("[[")
				str.WriteString(removeNonUtf8(node.FirstChild.Data))
				str.WriteString("]]")
			}
		}
	}
}

func removeNonUtf8(str string) string {
	var sb strings.Builder
	for _, r := range str {
		if utf8.ValidRune(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
