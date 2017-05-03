// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package ttndoc

type Package struct {
	TTNDoc *TTNDoc
	Files  []*File
	Name   string
}

func (d *TTNDoc) newPackage(name string) *Package {
	return &Package{TTNDoc: d, Name: name}
}

func (p *Package) AddFile(f *File) {
	p.Files = append(p.Files, f)
}

func (p Package) Document() bool {
	for _, file := range p.Files {
		if file.Document() {
			return true
		}
	}
	return false
}
