package models

import (
	"encoding/json"

	"github.com/reonardoleis/rinha-backend-2023/db"
)

type Person struct {
	ID        db.CustomUUID        `json:"id" db:"id"`
	Nickname  string               `json:"apelido" db:"nickname" validate:"required,max=32"`
	Name      string               `json:"nome" db:"name" validate:"required,max=100"`
	BirthDate db.CustomDate        `json:"nascimento" db:"birth_date" validate:"required,datetime=2006-01-02"`
	Stack     db.CustomStringSlice `json:"stack" db:"stack"`
}

func (p Person) JSON() []byte {
	out := `{"id":"` + string(p.ID) + `","apelido":"` + p.Nickname + `","nome":"` + p.Name + `","nascimento":"` + string(p.BirthDate) + `","stack":[`

	for i, stack := range p.Stack {
		out += `"` + stack + `"`
		if i < len(p.Stack)-1 {
			out += `,`
		}
	}

	out += `]}`
	return []byte(out)
}

func PersonListJSON(person []*Person) []byte {
	out := `[`
	for i, p := range person {
		out += string(p.JSON())
		if i < len(person)-1 {
			out += `,`
		}
	}
	out += `]`
	return []byte(out)
}

func (p *Person) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}
