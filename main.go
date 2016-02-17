package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

var client *http.Client

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

func filterNodes(node *html.Node) {
	switch node.Type {
	case html.ElementNode:
		var filteredAttr []html.Attribute
		for _, attr := range node.Attr {
			if attr.Key == "style" || attr.Key == "onclick" || attr.Key == "border" || attr.Key == "height" ||
				attr.Key == "width" {

			} else {
				filteredAttr = append(filteredAttr, attr)
			}
		}
		node.Attr = filteredAttr

		tagName := strings.Trim(node.Data, " \t\n")
		if tagName == "script" || tagName == "style" || tagName == "noscript" || tagName == "form" {
			if node.Parent != nil {
				node.Parent.RemoveChild(node)
			}
		} else if tagName == "meta" {
			log.Printf("%v", node)
			for _, attr := range node.Attr {
				if attr.Key == "http-equiv" || attr.Key == "charset" {

					if node.Parent != nil {
						node.Parent.RemoveChild(node)
					}
				}
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		filterNodes(c)
	}
}

func prettyPrint(w io.Writer, node *html.Node, level int, shouldProcess bool) {
	closingTag := ""
	switch node.Type {
	case html.ElementNode:
		tagName := strings.Trim(node.Data, " ")
		if tagName == "script" || tagName == "style" || tagName == "noscript" || tagName == "form" {
			shouldProcess = false
		} else {
			if tagName == "meta" || tagName == "img" || tagName == "link" {

			} else {
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
			w.Write(endTag)
		}

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
			prettyPrint(w, c, level+1, shouldProcess)
		}
	}

	if closingTag != "" {
		indent(w, level)
		w.Write(closingLt)
		w.Write([]byte(closingTag))
		w.Write(endTag)
	}
}

func toUtf8(iso8859 []byte) []rune {
	buf := make([]rune, len(iso8859))
	for i, b := range iso8859 {
		buf[i] = rune(b)
	}
	return buf
}

func main() {
	var url string
	flag.StringVar(&url, "url", "", "url to fetch")
	flag.Parse()
	if url == "" {
		flag.Usage()
		os.Exit(1)
	}

	client = &http.Client{
		Timeout: 10 * time.Second,
	}

	r, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()

	body, err := charset.NewReader(r.Body, "text/html")
	if err != nil {
		log.Fatal(err)
	}

	//x, _ := ioutil.ReadAll(body)
	//b := strings.NewReader(string(toUtf8(x)))
	//b := transform.NewReader(body, charmap.ISO8859_2.NewDecoder())
	root, err := html.Parse(body)
	if err != nil {
		log.Fatal(err)
	}
	filterNodes(root)
	prettyPrint(os.Stdout, root, 0, true)
	//html.Render(os.Stdout, root)

	/*
		tokenizer := html.NewTokenizer(body)
		for {
			t := tokenizer.Next()
			switch t {
			case html.ErrorToken:
				log.Fatal(tokenizer.Err())
			case html.StartTagToken:
				tag, _ := tokenizer.TagName()
				log.Printf("%s", tag)
			case html.EndTagToken:
				tag, _ := tokenizer.TagName()
				log.Printf("/%s", tag)
			}

		}
	*/

}
