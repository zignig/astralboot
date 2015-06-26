// questions for start up configuration
package main

//Set up the queries/
// functions in questions.go

func (c *Config) Setup() (q *qTree) {
	q = &qTree{
		title: "Configure Astralboot",
		questions: []Asker{
			&yesNoQuestion{text: "Enable Spawn", deflt: true},
			&yesNoQuestion{text: "test question 2"},
			&ipAddrQuestion{text: "IP address", ip: "test"},
		},
	}
	return
}
