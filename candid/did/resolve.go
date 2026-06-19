package did

import (
	"fmt"
	"os"
	"path/filepath"
)

// ParseDIDFile parses the .did file at path and resolves its import and
// `import service` declarations transitively, relative to the importing file.
//
// Type definitions from imported files are merged into the returned Description.
// Plain `import` brings types only. `import service` additionally merges the
// imported file's service methods into the importing file's service. On a method
// name clash the importing file wins.
func ParseDIDFile(path string) (*Description, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return resolveFile(abs)
}

type resolved struct {
	desc        *Description
	wantService bool
}

func resolveFile(abs string) (*Description, error) {
	desc, err := parseFile(abs)
	if err != nil {
		return nil, err
	}
	abs = canonical(abs)

	// Phase 1: walk the import graph once, recording per file the union of how it
	// was imported (plain vs service). onStack breaks cycles, and discovered
	// memoizes so each file is parsed once and a diamond does not re-read.
	discovered := map[string]*resolved{}
	onStack := map[string]bool{}
	var order []string

	var walk func(path string, desc *Description, asService bool) error
	walk = func(path string, desc *Description, asService bool) error {
		if e, ok := discovered[path]; ok {
			e.wantService = e.wantService || asService
			return nil
		}
		if onStack[path] {
			return nil
		}
		onStack[path] = true
		dir := filepath.Dir(path)
		for _, def := range desc.Definitions {
			imp, ok := def.(Import)
			if !ok {
				continue
			}
			target := imp.Text
			if !filepath.IsAbs(target) {
				target = filepath.Join(dir, target)
			}
			sub, err := parseFile(target)
			if err != nil {
				return err
			}
			if err := walk(canonical(target), sub, imp.Service); err != nil {
				return err
			}
		}
		onStack[path] = false
		discovered[path] = &resolved{desc: desc, wantService: asService}
		order = append(order, path)
		return nil
	}
	if err := walk(abs, desc, false); err != nil {
		return nil, err
	}

	// Phase 2: merge in post-order so dependencies precede dependents. The entry
	// file's own service merges first so it wins on method clashes, and its types
	// come last in `order`, so seed them first to keep the entry file authoritative.
	out := &Description{}
	for _, s := range discovered[abs].desc.Services {
		out.mergeService(s)
	}
	for i := len(order) - 1; i >= 0; i-- {
		r := discovered[order[i]]
		for _, def := range r.desc.Definitions {
			if t, ok := def.(Type); ok {
				if err := out.addType(t); err != nil {
					return nil, err
				}
			}
		}
		if order[i] != abs && r.wantService {
			for _, s := range r.desc.Services {
				out.mergeService(s)
			}
		}
	}
	return out, nil
}

// canonical resolves symlinks so the same file reached via different paths maps
// to one key. Callers must have confirmed the path exists (EvalSymlinks errors
// otherwise), and on any failure it falls back to the lexical path.
func canonical(path string) string {
	if c, err := filepath.EvalSymlinks(path); err == nil {
		return c
	}
	return path
}

func parseFile(abs string) (*Description, error) {
	raw, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	desc, err := ParseDID([]rune(string(raw)))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", abs, err)
	}
	return desc, nil
}

func (d *Description) addType(t Type) error {
	for _, def := range d.Definitions {
		e, ok := def.(Type)
		if !ok || e.Id != t.Id {
			continue
		}
		if e.Data.String() != t.Data.String() {
			return fmt.Errorf("conflicting definitions for type %q", t.Id)
		}
		return nil
	}
	d.Definitions = append(d.Definitions, t)
	return nil
}

// mergeService folds s into the single composed service, importer-wins on clashes.
func (d *Description) mergeService(s Service) {
	if len(d.Services) == 0 {
		d.Services = append(d.Services, Service{ID: s.ID})
	}
	dst := &d.Services[0]
	have := map[string]bool{}
	for _, m := range dst.Methods {
		have[m.Name] = true
	}
	for _, m := range s.Methods {
		if have[m.Name] {
			continue
		}
		have[m.Name] = true
		dst.Methods = append(dst.Methods, m)
	}
}
