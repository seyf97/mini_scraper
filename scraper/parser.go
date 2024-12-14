package scraper

import (
	"golang.org/x/net/html"
)

// Returns the text from a title element, if it exists
func get_title_data(n *html.Node) string {
	if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
		return n.FirstChild.Data
	}
	return ""
}

// BFS to find the title
func get_title(root *html.Node) string {
	queue := []*html.Node{root}

	// Queue of each node
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.Type == html.ElementNode && current.Data == "title" {
			title := get_title_data(current)
			if title != "" {
				return title
			}
		}

		for child := current.FirstChild; child != nil; child = child.NextSibling {
			queue = append(queue, child)
		}
	}

	return ""
}
