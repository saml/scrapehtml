package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

var lt = []byte("<")
var gt = []byte(">")
var closingLt = []byte("</")
var endTag = []byte(">\n")
var nl = []byte("\n")
var sp = []byte(" ")
var eq = []byte("=")
var comma = []byte(", ")
var openAttr = []byte(`="`)
var closeAttr = []byte(`"`)
var doctype = []byte("<!doctype html>\n")
var charsetMeta = []byte("<meta charset=\"utf-8\">\n")

func printAttr(w io.Writer, attr *html.Attribute) {
	w.Write(sp)
	w.Write([]byte(attr.Key))
	w.Write(openAttr)
	w.Write([]byte(attr.Val))
	w.Write(closeAttr)
}

func indent(w io.Writer, level int) {
	w.Write([]byte(strings.Repeat("  ", level)))
}

var charsetKeys = map[string]interface{}{
	"http-equiv": true,
	"charset":    true,
}

func hasKey(attributes []html.Attribute, keys map[string]interface{}) bool {
	for _, attr := range attributes {
		_, ok := keys[attr.Key]
		if ok {
			return true
		}

	}
	return false
}

func prettyPrint(w io.Writer, parentTag string, node *html.Node, level int, shouldProcess bool) {
	closingTag := ""
	switch node.Type {
	case html.ElementNode:
		tagName := strings.Trim(node.Data, " \t\n")
		if tagName == "script" || tagName == "style" || tagName == "noscript" || tagName == "form" {
			shouldProcess = false
		} else if !(tagName == "meta") || !hasKey(node.Attr, charsetKeys) {

			if !(tagName == "meta" || tagName == "img" || tagName == "link" || tagName == "br" ||
				tagName == "hr") {
				closingTag = tagName
			}
			indent(w, level)
			w.Write(lt)
			w.Write([]byte(tagName))
			for _, attr := range node.Attr {
				if attr.Key == "style" || attr.Key == "onclick" || attr.Key == "border" ||
					attr.Key == "height" || attr.Key == "width" {
				} else {
					printAttr(w, &attr)
				}
			}
			w.Write(gt)
			if !(tagName == "a" || tagName == "span" || tagName == "em" ||
				tagName == "string" || tagName == "i" || tagName == "title") {
				w.Write(nl)
			} else {
				w.Write(nl)
			}

			if tagName == "head" {
				indent(w, level+1)
				w.Write(charsetMeta) // charset utf-8 first
			}
		}
		level++
	case html.DoctypeNode:
		w.Write(doctype)
	case html.TextNode:
		txt := strings.Trim(node.Data, " \n\t")
		if txt != "" {
			indent(w, level+1)
			w.Write([]byte(txt))
			w.Write(nl)
		}
	}

	if shouldProcess {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			prettyPrint(w, closingTag, c, level, shouldProcess)
		}
	}

	if closingTag != "" {
		indent(w, level-1)
		w.Write(closingLt)
		w.Write([]byte(closingTag))
		w.Write(endTag)
	}
}

func ParseHTML(r io.Reader, cs string) (*html.Node, error) {
	var err error
	if cs == "" {
		// attempt to guess the charset of the HTML document
		r, err = charset.NewReader(r, "")
		if err != nil {
			return nil, err
		}
	} else {
		// let the user specify the charset
		e, name := charset.Lookup(cs)
		if name == "" {
			return nil, fmt.Errorf("'%s' is not a valid charset", cs)
		}
		r = transform.NewReader(r, e.NewDecoder())
	}
	return html.Parse(r)
}

func main() {
	root, err := ParseHTML(os.Stdin, "")
	if err != nil {
		log.Fatal(err)
	}

	prettyPrint(os.Stdout, "", root, 0, true)
}
