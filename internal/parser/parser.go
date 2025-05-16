package parser

import (
	"bufio"
	"fmt"
	"log"
	"regexp"
	"strings"

	"mermaid-icepanel/internal/api"
)

// FileReader provides an interface for reading files.
type FileReader interface {
	ReadFile(path string) (ReadCloser, error)
}

// DefaultFileReader reads files from the filesystem.
type DefaultFileReader struct{}

// ReadFile reads a file from the specified path.
func (r *DefaultFileReader) ReadFile(path string) (ReadCloser, error) {
	fileReader := &OsFileReader{}
	return fileReader.ReadFile(path)
}

// ---------- regex definitions ----------.
var (
	rePerson   = regexp.MustCompile(`^Person\(([^,]+),\s*"([^"]+)"(?:,\s*"([^"]+)")?\)`)
	reSystem   = regexp.MustCompile(`^System(_Ext|Db)?\(([^,]+),\s*"([^"]+)"(?:,\s*"([^"]+)")?\)`)
	reBound    = regexp.MustCompile(`^System_Boundary\(([^,]+),\s*"([^"]+)"\)`)
	reRel      = regexp.MustCompile(`^(Bi)?Rel\(([^,]+),\s*([^,]+),\s*"([^"]+)"\)`)
	reEndBound = regexp.MustCompile(`^}`)
)

// ---------- parser ----------.
type mermaidParser struct {
	objs  map[string]*api.Object
	conns []*api.Connection
	idSeq int
}

func newMermaidParser() *mermaidParser {
	return &mermaidParser{objs: make(map[string]*api.Object)}
}

func (p *mermaidParser) nextHandle() string {
	p.idSeq++
	return fmt.Sprintf("h%04d", p.idSeq)
}

func slug(id string) string {
	return strings.ToLower(strings.ReplaceAll(id, " ", "-"))
}

func (p *mermaidParser) addObj(id, name, desc, typ string) {
	if _, ok := p.objs[id]; ok {
		return
	}
	p.objs[id] = &api.Object{
		Handle: slug(id),
		Name:   name,
		Desc:   desc,
		Type:   typ,
	}
}

func (p *mermaidParser) addConn(from, to, label string, bidi bool) {
	h := p.nextHandle()
	p.conns = append(p.conns, &api.Connection{Handle: h, From: slug(from), To: slug(to), Label: label})
	if bidi {
		h2 := p.nextHandle()
		p.conns = append(p.conns, &api.Connection{Handle: h2, From: slug(to), To: slug(from), Label: label})
	}
}

// ParseMermaid parses a Mermaid file and returns an IcePanel diagram.
func ParseMermaid(fileReader FileReader, path string) (*api.Diagram, error) {
	f, err := fileReader.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", path, err)
	}

	defer func() {
		if cerr := f.Close(); cerr != nil {
			log.Printf("Error closing file: %v", cerr)
		}
	}()

	p := newMermaidParser()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if m := rePerson.FindStringSubmatch(line); m != nil {
			p.addObj(m[1], m[2], m[3], "actor")
			continue
		}
		if m := reSystem.FindStringSubmatch(line); m != nil {
			switch m[1] { // suffix group
			case "_Ext":
				p.addObj(m[2], m[3], m[4], "system")
				p.objs[m[2]].Props = map[string]interface{}{"external": true}
			case "Db":
				p.addObj(m[2], m[3], m[4], "store")
			default:
				p.addObj(m[2], m[3], m[4], "system")
			}
			continue
		}
		if m := reBound.FindStringSubmatch(line); m != nil {
			p.addObj(m[1], m[2], "", "group")
			continue
		}
		if m := reRel.FindStringSubmatch(line); m != nil {
			bidi := m[1] == "Bi"
			p.addConn(m[2], m[3], m[4], bidi)
			continue
		}
		if reEndBound.MatchString(line) {
			// ignore
			continue
		}
		// ignore unknown lines (layout etc.)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Build diagram
	d := &api.Diagram{
		Name:        "Imported Diagram",
		Type:        "app-diagram",
		Objects:     make([]*api.Object, 0, len(p.objs)),
		Connections: p.conns,
	}
	for _, o := range p.objs {
		d.Objects = append(d.Objects, o)
	}
	return d, nil
}
