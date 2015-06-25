// Questions for initial config
package main

import (
	"fmt"
)

var Configurate = &qTree{
	title: "Configure Astralboot",
	questions: []Asker{
		&yesNoQuestion{text: "test question"},
		&yesNoQuestion{text: "test question 2"},
	},
}

type Asker interface {
	Ask() (i interface{})
}

type qTree struct {
	title     string
	questions []Asker
	finished  bool
}

func (q *qTree) Run(c interface{}) {
	fmt.Println(q.title)
	for _, i := range q.questions {
		i.Ask()
	}
}

// base question
type q struct {
	text     string
	complete bool
}

// Yes or No question
type yesNoQuestion struct {
	q
	text     string
	detail   string
	complete bool
	val      bool
}

func (q *yesNoQuestion) Ask() (v interface{}) {
	fmt.Println(q.text)
	fmt.Println("")
	return true
}

// IP address
type ipAddrQuestion struct {
	q
}
