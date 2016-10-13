package components

import (
	"github.com/bep/gr"
	"github.com/bep/gr/el"
	"github.com/bep/grouter"
)

type Nav struct {
	*gr.This
	Pages
	Brand string
}

type Pages map[string]Page

type Page struct {
	Route           string
	DropdownOptions []DropdownOption
	ClassEndpoint   string
	AssetEndpoint   string
}

// Implements the Renderer interface.
func (c Nav) Render() gr.Component {

	links := el.UnorderedList(
		gr.CSS("nav-menu", "nav-pills", "nav-stacked"),
	)

	for name, page := range c.Pages {
		if page.Route != "/" {
			c.createLinkListItem(page.Route, name).Modify(links)
		}
	}

	elem := el.Div(gr.CSS("nav-wrapper"),
		el.ListItem(
			gr.CSS("nav-brand"),
			el.Italic(gr.CSS("fa", "fa-cogs")),
			gr.Text(" "),
			grouter.Link("/", c.Brand),
		),
		links,
	)

	return elem
}

func (c Nav) createLinkListItem(path, title string) gr.Modifier {
	return el.ListItem(
		grouter.MarkIfActive(c.Props(), path),
		grouter.Link(path, title),
	)
}
