// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package ttndoc

import "fmt"
import "strings"

type Proto struct {
	TTNDoc *TTNDoc
	File   *File
	Name   string
	Loc    string
}

func (d *TTNDoc) newProto(name string, loc string) Proto {
	return Proto{TTNDoc: d, Name: name, Loc: loc}
}

func (p Proto) String() string {
	return p.Name
}

func (p Proto) Comment() string {
	comment := p.TTNDoc.GetComment(fmt.Sprintf("%s:%s", p.File.Name, p.Loc))
	firstLine := strings.Index(comment, ".")
	if strings.HasPrefix(p.File.Path, "github.com/TheThingsNetwork") || firstLine == -1 {
		return comment
	}
	return comment[:firstLine+1]
}
