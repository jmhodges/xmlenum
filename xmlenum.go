package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"xml"
)

type TagMap map[string]TagMap

func main() {
	log.SetPrefix("")
	log.SetFlags(0)
	flag.Parse()
	if flag.NArg() < 2 {
		log.Fatalf("usage: xmlenum FIRST_ELEMENT_NAME FILES*\n")
	}

	firstElementName := flag.Args()[0]
	filepaths := flag.Args()[1:]
	files := make([]*os.File, len(filepaths))
	for i, fp := range filepaths {
		f, err := os.Open(fp)
		if err != nil {
			log.Fatalf("Couldn't open %s: %v\n", fp, err)
		}
		files[i] = f
		defer f.Close()
	}

	toplevel := TagMap{}

	for i, f := range files {
		p := xml.NewParser(f)

		err := start(p, firstElementName, toplevel)
		if err != nil && err != os.EOF {
			log.Fatalf("Couldn't parse %s: %v\n", filepaths[i], err)
		}
	}
	sortedPrint(toplevel, 0)
}

func start(p *xml.Parser, name string, m TagMap) os.Error {
	for {
		tok, err := p.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == name {
				err = recurse(p, name, m)
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func recurse(p *xml.Parser, name string, m TagMap) os.Error {
	hasRecursed := false
	for {
		tok, err := p.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			hasRecursed = true
			if m[name] == nil {
				m[name] = TagMap{}
			}
			err = recurse(p, t.Name.Local, m[name])
			if err != nil {
				return err
			}
		case xml.EndElement:
			// If hasRecursed stays false, we are in a tag that only contained
			// text, no child tags. We use this instead of `case xml.CharData`
			// because CharData can pop up as a token even between
			// StartElements, etc.
			if !hasRecursed {
				m[name] = nil
			}

			// If ending the element we entered recurse for, return
			if t.Name.Local == name {
				return nil
			}
		}
	}
	return nil
}

// Print tags that contain only text first, then tags with children.
func sortedPrint(m TagMap, indent int) {
	simple := make([]string, 0, len(m))
	nested := make([]string, 0, len(m))
	for k, v := range m {
		if v == nil {
			simple = append(simple, k)
		} else {
			nested = append(nested, k)
		}
	}
	sort.Strings(simple)
	sort.Strings(nested)
	keys := append(simple, nested...)
	for _, k := range keys {
		fmt.Printf("%*s%s", indent, " ", k)
		v := m[k]
		if v != nil {
			fmt.Printf(": {\n")
			sortedPrint(v, indent+4)
			fmt.Printf("%*s}\n", indent, " ")
		} else {
			fmt.Printf("\n")
		}
	}
}
