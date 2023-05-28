package main

import "github.com/andreyvit/mvp/forms"

func saveFormButtonBar() forms.Child {
	return &forms.Wrapper{
		Template: "buttonbar",
		Child: forms.Children{
			&forms.Button{
				TagOpts: forms.TagOpts{},
				Title:   "Save Changes",
			},
		},
	}
}

var adminFormStyle = &forms.Style{
	Classes: map[string]string{
		"header": "text-xl font-bold text-gray-900",
	},
}

var horizontalFormStyle = &forms.Style{
	Templates: []forms.Subst{
		{For: "item", Use: "item-horiz"},
		{For: "group", Use: "group-horiz"},
	},
}

var verticalFormStyle = &forms.Style{
	Templates: []forms.Subst{
		{For: "item", Use: "item-vert"},
		{For: "group", Use: "group-vert"},
	},
}

var cardsFormStyle = &forms.Style{
	Templates: []forms.Subst{
		{For: "item", Use: "item-vert"},
		{For: "group", Use: "group-cards"},
	},
}
